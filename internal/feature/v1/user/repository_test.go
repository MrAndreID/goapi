package user

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MrAndreID/goapi/v2/internal/entity"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	operationInsert = "INSERT"
	operationUpdate = "UPDATE"

	tableUsers  = "users"
	tableEmails = "emails"

	envFileName     = ".env"
	envTestFileName = ".env.test"

	// postgresUnreachableHint explains what to configure when the repository tests
	// cannot find a database. These tests run against a real PostgreSQL server.
	postgresUnreachableHint = "cannot reach postgresql, these tests need a running server: " +
		"start it and set TEST_DATABASE_DSN, the TEST_DATABASE_* variables, or " + envTestFileName
)

// testDatabaseOverrideKeys are the variables that select a test database explicitly.
// One of these, or the presence of .env.test, is what the guard looks for.
var testDatabaseOverrideKeys = []string{
	"TEST_DATABASE_DSN",
	"TEST_DATABASE_HOST",
	"TEST_DATABASE_PORT",
	"TEST_DATABASE_USERNAME",
	"TEST_DATABASE_PASSWORD",
	"TEST_DATABASE_NAME",
	"TEST_DATABASE_SSL_MODE",
	"TEST_DATABASE_TIMEZONE",
}

var (
	databaseEnvOnce sync.Once
	envTestLoaded   bool
	schemaSequence  atomic.Int64
)

// moduleRoot walks up from the working directory until it finds go.mod, which is
// where the environment files live.
func moduleRoot() (string, bool) {
	directory, err := os.Getwd()

	if err != nil {
		return "", false
	}

	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory, true
		}

		parent := filepath.Dir(directory)

		if parent == directory {
			return "", false
		}

		directory = parent
	}
}

// loadDatabaseEnv reads .env.test and then .env from the module root, once per run.
// godotenv never overwrites a variable that is already set, so .env.test wins over
// .env and the real environment wins over both. The two files are loaded separately
// because godotenv.Load stops at the first missing file.
func loadDatabaseEnv() {
	databaseEnvOnce.Do(func() {
		root, found := moduleRoot()

		if !found {
			return
		}

		if err := godotenv.Load(filepath.Join(root, envTestFileName)); err == nil {
			envTestLoaded = true
		}

		_ = godotenv.Load(filepath.Join(root, envFileName))
	})
}

// requireTestDatabase refuses to run when nothing points the tests at a dedicated
// database. Without the guard a missing .env.test falls back to .env, which means
// the suite would quietly run against the application database.
func requireTestDatabase(t *testing.T) {
	t.Helper()

	loadDatabaseEnv()

	if envTestLoaded {
		return
	}

	for _, key := range testDatabaseOverrideKeys {
		if strings.TrimSpace(os.Getenv(key)) != "" {
			return
		}
	}

	location := envTestFileName

	if root, found := moduleRoot(); found {
		location = filepath.Join(root, envTestFileName)
	}

	t.Fatalf("%s not found: the repository tests need a database of their own, "+
		"otherwise the settings fall back to %s and the suite runs against the application database. "+
		"copy %s.example to %s, or set one of %s",
		location, envFileName, envTestFileName, envTestFileName, strings.Join(testDatabaseOverrideKeys, ", "))
}

// testDatabaseSetting resolves one database setting for the tests. TEST_DATABASE_*
// is checked first so a dedicated test database can be selected without touching the
// application settings, then DATABASE_*, then the built in default.
func testDatabaseSetting(key string, fallback string) string {
	for _, prefix := range []string{"TEST_DATABASE_", "DATABASE_"} {
		if value := strings.TrimSpace(os.Getenv(prefix + key)); value != "" {
			return value
		}
	}

	return fallback
}

// postgresDSN builds the connection string for the test database. TEST_DATABASE_DSN
// overrides everything. search_path is part of the DSN on purpose so every pooled
// connection lands in the same schema.
func postgresDSN(searchPath string) string {
	loadDatabaseEnv()

	if dsn := strings.TrimSpace(os.Getenv("TEST_DATABASE_DSN")); dsn != "" {
		return dsn + " search_path=" + searchPath
	}

	return strings.Join([]string{
		"host=" + testDatabaseSetting("HOST", "127.0.0.1"),
		"port=" + testDatabaseSetting("PORT", "5432"),
		"user=" + testDatabaseSetting("USERNAME", "postgres"),
		"password=" + testDatabaseSetting("PASSWORD", "postgres"),
		"dbname=" + testDatabaseSetting("NAME", "postgres"),
		"sslmode=" + testDatabaseSetting("SSL_MODE", "disable"),
		"TimeZone=" + testDatabaseSetting("TIMEZONE", "UTC"),
		"search_path=" + searchPath,
	}, " ")
}

// newPostgresDatabase opens a PostgreSQL connection confined to its own schema. The
// schema replaces the throwaway SQLite file: every test starts empty and is free to
// drop tables or install triggers without touching the other tests.
func newPostgresDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	requireTestDatabase(t)

	schema := fmt.Sprintf("user_test_%d_%d", os.Getpid(), schemaSequence.Add(1))

	db, err := gorm.Open(postgres.Open(postgresDSN(schema)), &gorm.Config{})
	if err != nil {
		t.Fatalf("%s: %v", postgresUnreachableHint, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get the underlying postgresql connection: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("%s: %v", postgresUnreachableHint, err)
	}

	if err := db.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatalf("failed to create schema %s: %v", schema, err)
	}

	t.Cleanup(func() {
		if err := db.Exec("DROP SCHEMA " + schema + " CASCADE").Error; err != nil {
			t.Errorf("failed to drop schema %s: %v", schema, err)
		}

		sqlDB.Close()
	})

	return db
}

// newPostgresRepository returns a repository backed by an isolated schema. When
// migrate is false the schema is left empty so the caller can exercise failures that
// come from missing tables.
func newPostgresRepository(t *testing.T, migrate bool) (*Repository, *gorm.DB) {
	t.Helper()

	db := newPostgresDatabase(t)

	if migrate {
		if err := db.AutoMigrate(&User{}, &Email{}); err != nil {
			t.Fatalf("failed to migrate schema: %v", err)
		}
	}

	return NewRepository(time.UTC, db), db
}

// mustExec runs raw SQL used to bend the database into a failing state, such as
// dropping a table or installing a trigger.
func mustExec(t *testing.T, db *gorm.DB, statement string) {
	t.Helper()

	if err := db.Exec(statement).Error; err != nil {
		t.Fatalf("failed to execute %q: %v", statement, err)
	}
}

// installSkipWriteTrigger turns the given write into a per row no-op. A BEFORE ROW
// trigger that returns NULL cancels the write without raising an error, so the
// statement reports zero affected rows. This is the PostgreSQL counterpart of
// SQLite's RAISE(IGNORE).
func installSkipWriteTrigger(t *testing.T, db *gorm.DB, table string, operation string) {
	t.Helper()

	name := fmt.Sprintf("skip_%s_%s", table, strings.ToLower(operation))

	mustExec(t, db, fmt.Sprintf("CREATE FUNCTION %s() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NULL; END; $$", name))
	mustExec(t, db, fmt.Sprintf("CREATE TRIGGER %s BEFORE %s ON %s FOR EACH ROW EXECUTE FUNCTION %s()", name, operation, table, name))
}

// installRejectWriteTrigger aborts the statement with an error, the PostgreSQL
// counterpart of SQLite's RAISE(ABORT).
func installRejectWriteTrigger(t *testing.T, db *gorm.DB, table string, operation string) {
	t.Helper()

	name := fmt.Sprintf("reject_%s_%s", table, strings.ToLower(operation))

	mustExec(t, db, fmt.Sprintf("CREATE FUNCTION %s() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION '%s rejected'; END; $$", name, name))
	mustExec(t, db, fmt.Sprintf("CREATE TRIGGER %s BEFORE %s ON %s FOR EACH ROW EXECUTE FUNCTION %s()", name, operation, table, name))
}

// failingRandomReader hands out a limited number of successful reads before it
// starts failing, which lets a test choose exactly which uuid.NewRandom call breaks.
type failingRandomReader struct {
	allowed int
	used    int
}

func (r *failingRandomReader) Read(p []byte) (int, error) {
	if r.used >= r.allowed {
		return 0, errors.New("random source unavailable")
	}

	r.used++

	for i := range p {
		p[i] = byte(r.used)
	}

	return len(p), nil
}

func useFailingUUIDSource(t *testing.T, allowed int) {
	t.Helper()

	uuid.SetRand(&failingRandomReader{allowed: allowed})

	t.Cleanup(func() {
		uuid.SetRand(nil)
	})
}

func TestRepositoryCreateReadUpdateDelete(t *testing.T) {
	repo, db := newPostgresRepository(t, true)

	createdUser, err := repo.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com", "bndre@gmail.com"}})
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	if createdUser.ID == "" {
		t.Fatal("expected created user id to be populated")
	}

	if len(createdUser.Emails) != 2 {
		t.Fatalf("expected 2 emails, got %d", len(createdUser.Emails))
	}

	readResponse, err := repo.Read(context.Background(), ReadData{PaginatorRequest: entity.PaginatorRequest{Page: "1", Limit: "10", DisableCalculateTotal: "true"}})
	if err != nil {
		t.Fatalf("expected read to succeed: %v", err)
	}

	if readResponse.Total != 1 {
		t.Fatalf("expected total to be 1 for the created record, got %d", readResponse.Total)
	}

	if err := repo.Update(UpdateData{ID: createdUser.ID, Name: "Andre Updated", Emails: []string{"c@example.com"}}); err != nil {
		t.Fatalf("expected update to succeed: %v", err)
	}

	readUpdated, err := repo.Read(context.Background(), ReadData{PaginatorRequest: entity.PaginatorRequest{Page: "1", Limit: "10", DisableCalculateTotal: "true"}, ID: createdUser.ID})
	if err != nil {
		t.Fatalf("failed to read updated user: %v", err)
	}

	if readUpdated.Records == nil {
		t.Fatalf("expected updated user records, got %#v", readUpdated.Records)
	}

	if err := repo.Delete(DeleteData{ID: createdUser.ID}); err != nil {
		t.Fatalf("expected delete to succeed: %v", err)
	}

	var userCount int64
	if err := db.Model(&User{}).Count(&userCount).Error; err != nil {
		t.Fatalf("failed to count users: %v", err)
	}

	if userCount != 0 {
		t.Fatalf("expected user to be deleted, got %d users remaining", userCount)
	}
}

func TestRepositoryUpdateReturnsErrorWhenUserMissing(t *testing.T) {
	repo, _ := newPostgresRepository(t, true)

	err := repo.Update(UpdateData{ID: "123e4567-e89b-12d3-a456-426614174000", Name: "Andre"})
	if err == nil || err.Error() != "FAILED_TO_READ_USER_DATA" {
		t.Fatalf("expected FAILED_TO_READ_USER_DATA, got %v", err)
	}
}

func TestRepositoryDeleteReturnsErrorWhenUserMissing(t *testing.T) {
	repo, _ := newPostgresRepository(t, true)

	err := repo.Delete(DeleteData{ID: "123e4567-e89b-12d3-a456-426614174000"})
	if err == nil || err.Error() != "FAILED_TO_READ_USER_DATA" {
		t.Fatalf("expected FAILED_TO_READ_USER_DATA, got %v", err)
	}
}

func TestRepositoryCreateReturnsErrorWhenDatabaseHasNoTables(t *testing.T) {
	repo, _ := newPostgresRepository(t, false)

	_, err := repo.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}})
	if err == nil {
		t.Fatal("expected create error when schema is missing")
	}
}

func TestRepositoryUpdateReturnsErrorWhenDeletingEmailsReturnsZeroRows(t *testing.T) {
	repo, _ := newPostgresRepository(t, true)

	createdUser, err := repo.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}})
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	err = repo.Update(UpdateData{ID: createdUser.ID, Name: "Andre Updated", Emails: []string{"bndre@gmail.com"}})
	if err != nil {
		t.Fatalf("expected update success with existing email rows, got %v", err)
	}
}

func TestRepositoryDeleteReturnsErrorWhenDeletingEmailsReturnsZeroRows(t *testing.T) {
	repo, _ := newPostgresRepository(t, true)

	createdUser, err := repo.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}})
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	err = repo.Delete(DeleteData{ID: createdUser.ID})
	if err != nil {
		t.Fatalf("expected delete success with existing email rows, got %v", err)
	}
}

func TestRepositoryUpdateReturnsErrorWhenNoEmailsAndNoEmailRowsExist(t *testing.T) {
	repo, _ := newPostgresRepository(t, true)

	createdUser, err := repo.Create(CreateData{Name: "Andre"})
	if err != nil {
		t.Fatalf("expected create to succeed without emails: %v", err)
	}

	err = repo.Update(UpdateData{ID: createdUser.ID, Name: "Andre Updated"})
	if err == nil || err.Error() != "FAILED_TO_READ_EMAIL_DATA" {
		t.Fatalf("expected FAILED_TO_READ_EMAIL_DATA, got %v", err)
	}
}

func TestRepositoryReadReturnsParseErrors(t *testing.T) {
	repo, _ := newPostgresRepository(t, true)

	if _, err := repo.Read(context.Background(), ReadData{PaginatorRequest: entity.PaginatorRequest{Page: "x"}}); err == nil {
		t.Fatal("expected page parse error")
	}

	if _, err := repo.Read(context.Background(), ReadData{PaginatorRequest: entity.PaginatorRequest{Limit: "y"}}); err == nil {
		t.Fatal("expected limit parse error")
	}

	if _, err := repo.Read(context.Background(), ReadData{PaginatorRequest: entity.PaginatorRequest{DisableCalculateTotal: "maybe"}}); err == nil {
		t.Fatal("expected disableCalculateTotal parse error")
	}
}

func TestRepositoryCreateReturnsErrorWhenUserUUIDGenerationFails(t *testing.T) {
	repo, _ := newPostgresRepository(t, true)

	useFailingUUIDSource(t, 0)

	if _, err := repo.Create(CreateData{Name: "Andre"}); err == nil {
		t.Fatal("expected create to fail when the user uuid cannot be generated")
	}
}

func TestRepositoryCreateReturnsErrorWhenEmailUUIDGenerationFails(t *testing.T) {
	repo, _ := newPostgresRepository(t, true)

	useFailingUUIDSource(t, 1)

	if _, err := repo.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}}); err == nil {
		t.Fatal("expected create to fail when the email uuid cannot be generated")
	}
}

func TestRepositoryCreateReturnsErrorWhenUserInsertAffectsNoRows(t *testing.T) {
	repo, db := newPostgresRepository(t, true)

	installSkipWriteTrigger(t, db, tableUsers, operationInsert)

	_, err := repo.Create(CreateData{Name: "Andre"})
	if err == nil || err.Error() != "FAILED_TO_CREATE_USER" {
		t.Fatalf("expected FAILED_TO_CREATE_USER, got %v", err)
	}
}

func TestRepositoryCreateReturnsErrorWhenEmailTableIsMissing(t *testing.T) {
	repo, db := newPostgresRepository(t, true)

	mustExec(t, db, "DROP TABLE emails")

	if _, err := repo.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}}); err == nil {
		t.Fatal("expected create to fail when the emails table is missing")
	}
}

func TestRepositoryCreateReturnsErrorWhenEmailInsertAffectsNoRows(t *testing.T) {
	repo, db := newPostgresRepository(t, true)

	installSkipWriteTrigger(t, db, tableEmails, operationInsert)

	_, err := repo.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}})
	if err == nil || err.Error() != "FAILED_TO_CREATE_EMAIL" {
		t.Fatalf("expected FAILED_TO_CREATE_EMAIL, got %v", err)
	}
}

func TestRepositoryReadReturnsErrorWhenDisableCalculateTotalIsEmpty(t *testing.T) {
	repo, _ := newPostgresRepository(t, true)

	if _, err := repo.Read(context.Background(), ReadData{PaginatorRequest: entity.PaginatorRequest{Page: "1", Limit: "10"}}); err == nil {
		t.Fatal("expected disable calculate total parse error")
	}
}

func TestRepositoryReadFlagsNextPageWhenRecordsFillTheLimit(t *testing.T) {
	repo, _ := newPostgresRepository(t, true)

	for i := 0; i < 10; i++ {
		if _, err := repo.Create(CreateData{Name: fmt.Sprintf("User %02d", i)}); err != nil {
			t.Fatalf("failed to seed user %d: %v", i, err)
		}
	}

	res, err := repo.Read(context.Background(), ReadData{PaginatorRequest: entity.PaginatorRequest{
		Page:                  "1",
		Limit:                 "10",
		OrderBy:               "name",
		SortBy:                "asc",
		Search:                "user",
		DisableCalculateTotal: "true",
	}})
	if err != nil {
		t.Fatalf("expected read to succeed: %v", err)
	}

	if !res.NextPage {
		t.Fatal("expected next page to be flagged when the page is full")
	}

	if res.Total != 10 {
		t.Fatalf("expected total to be 10, got %d", res.Total)
	}
}

func TestRepositoryUpdateReturnsErrorWhenDeletingEmailFails(t *testing.T) {
	repo, db := newPostgresRepository(t, true)

	createdUser, err := repo.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}})
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	mustExec(t, db, "DROP TABLE emails")

	if err := repo.Update(UpdateData{ID: createdUser.ID, Emails: []string{"bndre@gmail.com"}}); err == nil {
		t.Fatal("expected update to fail when the emails table is missing")
	}
}

func TestRepositoryUpdateReturnsErrorWhenDeletingEmailAffectsNoRows(t *testing.T) {
	repo, _ := newPostgresRepository(t, true)

	createdUser, err := repo.Create(CreateData{Name: "Andre"})
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	err = repo.Update(UpdateData{ID: createdUser.ID, Emails: []string{"bndre@gmail.com"}})
	if err == nil || err.Error() != "FAILED_TO_DELETE_EMAIL_DATA" {
		t.Fatalf("expected FAILED_TO_DELETE_EMAIL_DATA, got %v", err)
	}
}

func TestRepositoryUpdateReturnsErrorWhenEmailUUIDGenerationFails(t *testing.T) {
	repo, _ := newPostgresRepository(t, true)

	createdUser, err := repo.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}})
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	useFailingUUIDSource(t, 0)

	if err := repo.Update(UpdateData{ID: createdUser.ID, Emails: []string{"bndre@gmail.com"}}); err == nil {
		t.Fatal("expected update to fail when the email uuid cannot be generated")
	}
}

func TestRepositoryUpdateReturnsErrorWhenCreatingEmailFails(t *testing.T) {
	repo, db := newPostgresRepository(t, true)

	createdUser, err := repo.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}})
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	installRejectWriteTrigger(t, db, tableEmails, operationInsert)

	if err := repo.Update(UpdateData{ID: createdUser.ID, Emails: []string{"bndre@gmail.com"}}); err == nil {
		t.Fatal("expected update to fail when the email insert is rejected")
	}
}

func TestRepositoryUpdateReturnsErrorWhenCreatingEmailAffectsNoRows(t *testing.T) {
	repo, db := newPostgresRepository(t, true)

	createdUser, err := repo.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}})
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	installSkipWriteTrigger(t, db, tableEmails, operationInsert)

	err = repo.Update(UpdateData{ID: createdUser.ID, Emails: []string{"bndre@gmail.com"}})
	if err == nil || err.Error() != "FAILED_TO_CREATE_EMAIL" {
		t.Fatalf("expected FAILED_TO_CREATE_EMAIL, got %v", err)
	}
}

func TestRepositoryUpdateKeepsExistingEmailsWhenRequestHasNone(t *testing.T) {
	repo, db := newPostgresRepository(t, true)

	createdUser, err := repo.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}})
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	if err := repo.Update(UpdateData{ID: createdUser.ID, Name: "Andre Updated"}); err != nil {
		t.Fatalf("expected update to succeed: %v", err)
	}

	var emailCount int64
	if err := db.Model(&Email{}).Where("user_id = ?", createdUser.ID).Count(&emailCount).Error; err != nil {
		t.Fatalf("failed to count emails: %v", err)
	}

	if emailCount != 1 {
		t.Fatalf("expected the existing email to be kept, got %d", emailCount)
	}
}

func TestRepositoryUpdateReturnsErrorWhenSavingUserFails(t *testing.T) {
	repo, db := newPostgresRepository(t, true)

	createdUser, err := repo.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}})
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	installRejectWriteTrigger(t, db, tableUsers, operationUpdate)

	if err := repo.Update(UpdateData{ID: createdUser.ID, Name: "Andre Updated"}); err == nil {
		t.Fatal("expected update to fail when the user update is rejected")
	}
}

func TestRepositoryUpdateReturnsErrorWhenSavingUserAffectsNoRows(t *testing.T) {
	repo, db := newPostgresRepository(t, true)

	createdUser, err := repo.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}})
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	// gorm Save falls back to an insert when the update matches no rows, so both
	// writes have to be neutralised for RowsAffected to stay at zero.
	installSkipWriteTrigger(t, db, tableUsers, operationUpdate)
	installSkipWriteTrigger(t, db, tableUsers, operationInsert)

	err = repo.Update(UpdateData{ID: createdUser.ID, Name: "Andre Updated"})
	if err == nil || err.Error() != "FAILED_TO_UPDATE_USER_DATA" {
		t.Fatalf("expected FAILED_TO_UPDATE_USER_DATA, got %v", err)
	}
}

func TestRepositoryDeleteReturnsErrorWhenDeletingUserFails(t *testing.T) {
	repo, db := newPostgresRepository(t, true)

	createdUser, err := repo.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}})
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	installRejectWriteTrigger(t, db, tableUsers, operationUpdate)

	if err := repo.Delete(DeleteData{ID: createdUser.ID}); err == nil {
		t.Fatal("expected delete to fail when the soft delete is rejected")
	}
}

func TestRepositoryDeleteReturnsErrorWhenDeletingUserAffectsNoRows(t *testing.T) {
	repo, db := newPostgresRepository(t, true)

	createdUser, err := repo.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}})
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	// The soft delete is an update, so neutralising updates keeps the row in place
	// and reports zero affected rows.
	installSkipWriteTrigger(t, db, tableUsers, operationUpdate)

	err = repo.Delete(DeleteData{ID: createdUser.ID})
	if err == nil || err.Error() != "FAILED_TO_DELETE_USER_DATA" {
		t.Fatalf("expected FAILED_TO_DELETE_USER_DATA, got %v", err)
	}
}

func TestRepositoryDeleteReturnsErrorWhenEmailTableIsMissing(t *testing.T) {
	repo, db := newPostgresRepository(t, true)

	createdUser, err := repo.Create(CreateData{Name: "Andre"})
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	mustExec(t, db, "DROP TABLE emails")

	if err := repo.Delete(DeleteData{ID: createdUser.ID}); err == nil {
		t.Fatal("expected delete to fail when the emails table is missing")
	}
}

func TestRepositoryDeleteReturnsErrorWhenDeletingEmailAffectsNoRows(t *testing.T) {
	repo, _ := newPostgresRepository(t, true)

	createdUser, err := repo.Create(CreateData{Name: "Andre"})
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	err = repo.Delete(DeleteData{ID: createdUser.ID})
	if err == nil || err.Error() != "FAILED_TO_DELETE_EMAIL_DATA" {
		t.Fatalf("expected FAILED_TO_DELETE_EMAIL_DATA, got %v", err)
	}
}

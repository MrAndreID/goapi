package health_test

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MrAndreID/goapi/v2/internal/application/cache"
	messageBroker "github.com/MrAndreID/goapi/v2/internal/application/message_broker"
	objectStorage "github.com/MrAndreID/goapi/v2/internal/application/object_storage"
	. "github.com/MrAndreID/goapi/v2/internal/feature/health"

	"github.com/MrAndreID/gopackage/v2"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	envFileName     = ".env"
	envTestFileName = ".env.test"

	defaultConnectTimeoutSeconds = "5"

	postgresUnreachableHint = "cannot reach postgresql, this test needs a running server: " +
		"start it and set TEST_DATABASE_DSN, the TEST_DATABASE_* variables, or " + envTestFileName
)

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
)

// moduleRoot walks up from the working directory until it finds go.mod, where the
// environment files live.
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

// loadDatabaseEnv reads .env.test then .env from the module root, once per run.
// godotenv never overwrites a variable already set, so .env.test wins over .env
// and the real environment wins over both.
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

// requireTestDatabase refuses to run when nothing points the test at a dedicated
// database, so a missing .env.test cannot silently fall back to the application one.
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

	t.Skipf("no test database configured: set one of %s or provide %s",
		strings.Join(testDatabaseOverrideKeys, ", "), envTestFileName)
}

func testDatabaseSetting(key string, fallback string) string {
	for _, prefix := range []string{"TEST_DATABASE_", "DATABASE_"} {
		if value := strings.TrimSpace(os.Getenv(prefix + key)); value != "" {
			return value
		}
	}

	return fallback
}

func postgresDSN() string {
	loadDatabaseEnv()

	if dsn := strings.TrimSpace(os.Getenv("TEST_DATABASE_DSN")); dsn != "" {
		return withConnectTimeout(dsn)
	}

	return withConnectTimeout(strings.Join([]string{
		"host=" + testDatabaseSetting("HOST", "127.0.0.1"),
		"port=" + testDatabaseSetting("PORT", "5432"),
		"user=" + testDatabaseSetting("USERNAME", "postgres"),
		"password=" + testDatabaseSetting("PASSWORD", "postgres"),
		"dbname=" + testDatabaseSetting("NAME", "postgres"),
		"sslmode=" + testDatabaseSetting("SSL_MODE", "disable"),
		"TimeZone=" + testDatabaseSetting("TIMEZONE", "UTC"),
	}, " "))
}

func withConnectTimeout(dsn string) string {
	if strings.Contains(dsn, "connect_timeout") {
		return dsn
	}

	return dsn + " connect_timeout=" + testDatabaseSetting("CONNECT_TIMEOUT", defaultConnectTimeoutSeconds)
}

// TestRepositoryCheckDatabaseReportsHealthyWithoutDatabase covers the nil path: a
// repository built with USE_DATABASE=false reports healthy without touching a
// connection.
func TestRepositoryCheckDatabaseReportsHealthyWithoutDatabase(t *testing.T) {
	repo := NewRepository(nil, nil, nil, nil)

	status, err := repo.CheckDatabase(context.Background())
	if err != nil {
		t.Fatalf("expected no error when no database is wired, got %v", err)
	}

	if !status {
		t.Fatalf("expected healthy status when no database is wired")
	}
}

// TestRepositoryCheckDatabaseReportsHealthyWhenReachable covers the happy path
// against a real PostgreSQL server: a successful ping reports healthy.
func TestRepositoryCheckDatabaseReportsHealthyWhenReachable(t *testing.T) {
	requireTestDatabase(t)

	db, err := gorm.Open(postgres.Open(postgresDSN()), &gorm.Config{})
	if err != nil {
		t.Fatalf("%s: %v", postgresUnreachableHint, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get the underlying postgresql connection: %v", err)
	}

	t.Cleanup(func() { sqlDB.Close() })

	repo := NewRepository(db, nil, nil, nil)

	status, err := repo.CheckDatabase(context.Background())
	if err != nil {
		t.Fatalf("expected healthy against a reachable database, got %v", err)
	}

	if !status {
		t.Fatalf("expected healthy status against a reachable database")
	}
}

// invalidConnPool satisfies gorm's ConnPool interface but is not an *sql.DB, so
// gorm.DB.DB() returns ErrInvalidDB. None of its methods are ever called; they
// exist only to make the value assignable to Config.ConnPool.
type invalidConnPool struct{}

func (invalidConnPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, nil
}

func (invalidConnPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (invalidConnPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

func (invalidConnPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}

// TestRepositoryCheckDatabaseReportsErrorWhenUnderlyingDBUnavailable covers the
// branch where gorm.DB.DB() itself fails: a *gorm.DB whose ConnPool is not an
// *sql.DB makes DB() return ErrInvalidDB before any ping happens.
func TestRepositoryCheckDatabaseReportsErrorWhenUnderlyingDBUnavailable(t *testing.T) {
	db := &gorm.DB{Config: &gorm.Config{ConnPool: invalidConnPool{}}}

	repo := NewRepository(db, nil, nil, nil)

	status, err := repo.CheckDatabase(context.Background())
	if err == nil {
		t.Fatal("expected an error when the underlying database is invalid")
	}

	if status {
		t.Fatal("expected unhealthy status when the underlying database is invalid")
	}
}

// TestRepositoryCheckDatabaseReportsUnhealthyWhenUnreachable covers the ping-failure
// path: a connection to a port nothing listens on cannot ping, so the repository
// reports an error and false. DisableAutomaticPing lets Open succeed so the failure
// surfaces on PingContext, which is what CheckDatabase calls.
func TestRepositoryCheckDatabaseReportsUnhealthyWhenUnreachable(t *testing.T) {
	db, err := gorm.Open(postgres.Open("host=127.0.0.1 port=1 user=invalid dbname=invalid sslmode=disable connect_timeout=1"), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("failed to create unavailable database: %v", err)
	}

	repo := NewRepository(db, nil, nil, nil)

	status, err := repo.CheckDatabase(context.Background())
	if err == nil {
		t.Fatal("expected an error when the database is unreachable")
	}

	if status {
		t.Fatal("expected unhealthy status when the database is unreachable")
	}
}

// --- CheckCache ---

// TestRepositoryCheckCacheReportsHealthyWithoutCache covers the nil path: a
// repository built with USE_CACHE=false reports healthy without touching a client.
func TestRepositoryCheckCacheReportsHealthyWithoutCache(t *testing.T) {
	repo := NewRepository(nil, nil, nil, nil)

	status, err := repo.CheckCache(context.Background())
	if err != nil {
		t.Fatalf("expected no error when no cache is wired, got %v", err)
	}

	if !status {
		t.Fatalf("expected healthy status when no cache is wired")
	}
}

// TestRepositoryCheckCacheReportsUnhealthyWhenRedisUnreachable covers the Redis
// failure path: a client pointed at a dead port cannot ping.
func TestRepositoryCheckCacheReportsUnhealthyWhenRedisUnreachable(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:        deadAddr(t),
		DialTimeout: time.Second,
		ReadTimeout: time.Second,
	})

	t.Cleanup(func() { redisClient.Close() })

	repo := NewRepository(nil, &cache.CacheConnection{Redis: redisClient}, nil, nil)

	status, err := repo.CheckCache(context.Background())
	if err == nil {
		t.Fatal("expected an error when redis is unreachable")
	}

	if status {
		t.Fatal("expected unhealthy status when redis is unreachable")
	}
}

// TestRepositoryCheckCacheReportsUnhealthyWhenMemcachedUnreachable covers the
// Memcached failure path.
func TestRepositoryCheckCacheReportsUnhealthyWhenMemcachedUnreachable(t *testing.T) {
	client := memcache.New(deadAddr(t))
	client.Timeout = time.Second

	repo := NewRepository(nil, &cache.CacheConnection{Memcached: client}, nil, nil)

	status, err := repo.CheckCache(context.Background())
	if err == nil {
		t.Fatal("expected an error when memcached is unreachable")
	}

	if status {
		t.Fatal("expected unhealthy status when memcached is unreachable")
	}
}

// --- CheckMessageBroker ---

// TestRepositoryCheckMessageBrokerReportsHealthyWithoutBroker covers the nil path.
func TestRepositoryCheckMessageBrokerReportsHealthyWithoutBroker(t *testing.T) {
	repo := NewRepository(nil, nil, nil, nil)

	status, err := repo.CheckMessageBroker(context.Background())
	if err != nil {
		t.Fatalf("expected no error when no message broker is wired, got %v", err)
	}

	if !status {
		t.Fatalf("expected healthy status when no message broker is wired")
	}
}

// TestRepositoryCheckMessageBrokerReportsUnhealthyWhenRabbitMQConnectionNil covers
// the RabbitMQ path where the connection handle is missing: treated as closed.
func TestRepositoryCheckMessageBrokerReportsUnhealthyWhenRabbitMQConnectionNil(t *testing.T) {
	repo := NewRepository(nil, nil, &messageBroker.MessageBrokerConnection{
		RabbitMQ: &messageBroker.RabbitMQConnection{Connection: nil},
	}, nil)

	status, err := repo.CheckMessageBroker(context.Background())
	if err != nil {
		t.Fatalf("expected no error for a closed rabbitmq handle, got %v", err)
	}

	if status {
		t.Fatal("expected unhealthy status when the rabbitmq connection is nil")
	}
}

// TestRepositoryCheckMessageBrokerReportsUnhealthyWhenKafkaClosed covers the Kafka
// failure path: a connection closed before the check cannot reach the controller.
func TestRepositoryCheckMessageBrokerReportsUnhealthyWhenKafkaClosed(t *testing.T) {
	kafkaConn := dialClosedKafka(t)

	repo := NewRepository(nil, nil, &messageBroker.MessageBrokerConnection{Kafka: kafkaConn}, nil)

	status, err := repo.CheckMessageBroker(context.Background())
	if err == nil {
		t.Fatal("expected an error when the kafka connection is closed")
	}

	if status {
		t.Fatal("expected unhealthy status when the kafka connection is closed")
	}
}

// --- CheckObjectStorage ---

// TestRepositoryCheckObjectStorageReportsHealthyWithoutObjectStorage covers the nil
// path.
func TestRepositoryCheckObjectStorageReportsHealthyWithoutObjectStorage(t *testing.T) {
	repo := NewRepository(nil, nil, nil, nil)

	status, err := repo.CheckObjectStorage(context.Background())
	if err != nil {
		t.Fatalf("expected no error when no object storage is wired, got %v", err)
	}

	if !status {
		t.Fatalf("expected healthy status when no object storage is wired")
	}
}

// TestRepositoryCheckObjectStorageReportsHealthyWhenSeaweedFSReturnsOK covers the
// SeaweedFS happy path against a local test server that answers 200.
func TestRepositoryCheckObjectStorageReportsHealthyWhenSeaweedFSReturnsOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Cleanup(server.Close)

	repo := NewRepository(nil, nil, nil, &objectStorage.ObjectStorageConnection{
		SeaweedFS: &gopackage.SeaweedFSData{URL: server.URL},
	})

	status, err := repo.CheckObjectStorage(context.Background())
	if err != nil {
		t.Fatalf("expected healthy against a reachable seaweedfs, got %v", err)
	}

	if !status {
		t.Fatalf("expected healthy status against a reachable seaweedfs")
	}
}

// TestRepositoryCheckObjectStorageReportsUnhealthyWhenSeaweedFSReturnsNon200 covers
// the SeaweedFS branch where the server answers but with a non-200 status.
func TestRepositoryCheckObjectStorageReportsUnhealthyWhenSeaweedFSReturnsNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	t.Cleanup(server.Close)

	repo := NewRepository(nil, nil, nil, &objectStorage.ObjectStorageConnection{
		SeaweedFS: &gopackage.SeaweedFSData{URL: server.URL},
	})

	status, err := repo.CheckObjectStorage(context.Background())
	if err != nil {
		t.Fatalf("expected no transport error for a non-200 response, got %v", err)
	}

	if status {
		t.Fatal("expected unhealthy status when seaweedfs returns a non-200 status")
	}
}

// TestRepositoryCheckObjectStorageReportsUnhealthyWhenSeaweedFSUnreachable covers
// the SeaweedFS transport-failure path: the server is closed before the request.
func TestRepositoryCheckObjectStorageReportsUnhealthyWhenSeaweedFSUnreachable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := server.URL
	server.Close()

	repo := NewRepository(nil, nil, nil, &objectStorage.ObjectStorageConnection{
		SeaweedFS: &gopackage.SeaweedFSData{URL: url},
	})

	status, err := repo.CheckObjectStorage(context.Background())
	if err == nil {
		t.Fatal("expected an error when seaweedfs is unreachable")
	}

	if status {
		t.Fatal("expected unhealthy status when seaweedfs is unreachable")
	}
}

// TestRepositoryCheckObjectStorageReportsUnhealthyWhenSeaweedFSURLInvalid covers the
// branch where building the request itself fails: a URL with a control character
// cannot be parsed into an *http.Request, so no network call is attempted.
func TestRepositoryCheckObjectStorageReportsUnhealthyWhenSeaweedFSURLInvalid(t *testing.T) {
	repo := NewRepository(nil, nil, nil, &objectStorage.ObjectStorageConnection{
		SeaweedFS: &gopackage.SeaweedFSData{URL: "http://\x7f-invalid"},
	})

	status, err := repo.CheckObjectStorage(context.Background())
	if err == nil {
		t.Fatal("expected an error when the seaweedfs url cannot be turned into a request")
	}

	if status {
		t.Fatal("expected unhealthy status when the seaweedfs url is invalid")
	}
}

// TestRepositoryCheckObjectStorageReportsUnhealthyWhenMinioUnreachable covers the
// MinIO failure path: a client pointed at a dead port cannot list buckets.
func TestRepositoryCheckObjectStorageReportsUnhealthyWhenMinioUnreachable(t *testing.T) {
	minioClient, err := minio.New(deadAddr(t), &minio.Options{})
	if err != nil {
		t.Fatalf("failed to build minio client: %v", err)
	}

	repo := NewRepository(nil, nil, nil, &objectStorage.ObjectStorageConnection{Minio: minioClient})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	status, err := repo.CheckObjectStorage(ctx)
	if err == nil {
		t.Fatal("expected an error when minio is unreachable")
	}

	if status {
		t.Fatal("expected unhealthy status when minio is unreachable")
	}
}

// deadAddr returns a host:port that accepts no connections. It binds a listener to
// obtain a free port and closes it immediately, so nothing is listening there.
func deadAddr(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve a port: %v", err)
	}

	address := listener.Addr().String()

	if err := listener.Close(); err != nil {
		t.Fatalf("failed to release the reserved port: %v", err)
	}

	return address
}

// dialClosedKafka returns a *kafka.Conn wrapping a TCP connection that is already
// closed, so any request such as Controller() fails immediately. A local listener
// accepts one connection to make the dial succeed, then everything is torn down.
func dialClosedKafka(t *testing.T) *kafka.Conn {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start a listener: %v", err)
	}

	t.Cleanup(func() { listener.Close() })

	rawConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial the listener: %v", err)
	}

	kafkaConn := kafka.NewConn(rawConn, "health", 0)

	if err := kafkaConn.Close(); err != nil {
		t.Fatalf("failed to close the kafka connection: %v", err)
	}

	return kafkaConn
}

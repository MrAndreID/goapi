package database

import (
	"errors"
	"strconv"
	"strings"
	"time"

	gosqlmysql "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	Connection     string
	Host           string
	Port           string
	Username       string
	Password       string
	Name           string
	SSLMode        string
	ParseTime      string
	Charset        string
	Timezone       string
	ConnectTimeout string
}

func New(database *Database, debug bool) (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
	)

	switch database.Connection {
	case "postgresql":
		db, err = database.PostgreSQL()
	case "mysql":
		db, err = database.MySQL()
	default:
		err = errors.New("Database Connection Not Found")
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   "internal.application.database.main.New.01",
			"error": err.Error(),
		}).Error("failed to connect database")

		return nil, err
	}

	if debug {
		return db.Debug(), nil
	}

	return db, nil
}

func (database *Database) PostgreSQL() (*gorm.DB, error) {
	dsn := strings.Join([]string{
		"host=" + quotePostgresValue(database.Host),
		"user=" + quotePostgresValue(database.Username),
		"password=" + quotePostgresValue(database.Password),
		"dbname=" + quotePostgresValue(database.Name),
		"port=" + quotePostgresValue(database.Port),
		"sslmode=" + quotePostgresValue(database.SSLMode),
		"TimeZone=" + quotePostgresValue(database.Timezone),
		"connect_timeout=" + quotePostgresValue(database.ConnectTimeout),
	}, " ")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   "internal.application.database.main.PostgreSQL.01",
			"error": err.Error(),
		}).Error("failed to connect postgresql database")

		return nil, err
	}

	return db, nil
}

func (database *Database) MySQL() (*gorm.DB, error) {
	var tag string = "internal.application.database.main.MySQL."

	location, err := time.LoadLocation(database.Timezone)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.Error(),
		}).Error("failed to load location for mysql timezone")

		return nil, err
	}

	parseTime, err := strconv.ParseBool(database.ParseTime)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to parse mysql parse time flag")

		return nil, err
	}

	mysqlConfig := gosqlmysql.NewConfig()
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = database.Host + ":" + database.Port
	mysqlConfig.User = database.Username
	mysqlConfig.Passwd = database.Password
	mysqlConfig.DBName = database.Name
	mysqlConfig.Collation = ""
	mysqlConfig.ParseTime = parseTime
	mysqlConfig.Loc = location
	mysqlConfig.Params = map[string]string{
		"charset": database.Charset,
	}
	mysqlConfig.Timeout = time.Duration(connectTimeoutSeconds(database.ConnectTimeout)) * time.Second

	db, err := gorm.Open(mysql.Open(mysqlConfig.FormatDSN()), &gorm.Config{})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "03",
			"error": err.Error(),
		}).Error("failed to connect mysql database")

		return nil, err
	}

	return db, nil
}

func quotePostgresValue(value string) string {
	if value != "" && !strings.ContainsAny(value, " '\\") {
		return value
	}

	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)

	return "'" + escaped + "'"
}

func connectTimeoutSeconds(value string) int {
	seconds, err := strconv.Atoi(value)

	if err != nil {
		return 0
	}

	return seconds
}

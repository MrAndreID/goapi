package databases

import (
	"errors"
	"strings"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	Connection string
	Host       string
	Port       string
	Username   string
	Password   string
	Name       string
	SSLMode    string
	ParseTime  string
	Charset    string
	Timezone   string
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
			"tag":   "Databases.Main.New.01",
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
	dsn := "host=" + database.Host + " user=" + database.Username + " password=" + database.Password + " dbname=" + database.Name + " port=" + database.Port + " sslmode=" + database.SSLMode + " TimeZone=" + database.Timezone

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   "Databases.Main.PostgreSQL.01",
			"error": err.Error(),
		}).Error("failed to connect postgresql database")

		return nil, err
	}

	return db, nil
}

func (database *Database) MySQL() (*gorm.DB, error) {
	timezone := strings.Replace(database.Timezone, "/", "%2F", -1)

	dsn := database.Username + ":" + database.Password + "@tcp(" + database.Host + ":" + database.Port + ")/" + database.Name + "?charset=" + database.Charset + "&parseTime=" + database.ParseTime + "&loc=" + timezone

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   "Databases.Main.MySQL.01",
			"error": err.Error(),
		}).Error("failed to connect mysql database")

		return nil, err
	}

	return db, nil
}

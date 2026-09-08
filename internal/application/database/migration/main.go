package main

import (
	"flag"
	"fmt"

	"github.com/MrAndreID/goapi/v2/internal/application/config"
	"github.com/MrAndreID/goapi/v2/internal/application/database"
	"github.com/MrAndreID/goapi/v2/internal/feature/v1/user"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type table struct {
	Name  string
	Model any
}

var tables []table = []table{
	{Name: "users", Model: &user.User{}},
	{Name: "emails", Model: &user.Email{}},
}

func main() {
	var tag string = "internal.application.database.migration.main.main."

	cfg, err := config.New()

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.Error(),
		}).Error("failed to initiate configuration")

		return
	}

	var dbConnection *gorm.DB

	if !cfg.UseDatabase {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": "The Database is Not Yet Used",
		}).Error("failed to migrate")

		return
	} else {
		dbConnection, err = database.New(&database.Database{
			Connection:     cfg.DatabaseConnection,
			Host:           cfg.DatabaseHost,
			Port:           cfg.DatabasePort,
			Username:       cfg.DatabaseUsername,
			Password:       cfg.DatabasePassword,
			Name:           cfg.DatabaseName,
			SSLMode:        cfg.DatabaseSSLMode,
			ParseTime:      cfg.DatabaseParseTime,
			Charset:        cfg.DatabaseCharset,
			Timezone:       cfg.DatabaseTimezone,
			ConnectTimeout: cfg.DatabaseConnectTimeout,
		}, cfg.AppDebug)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "03",
				"error": err.Error(),
			}).Error("failed to connect database")

			return
		}
	}

	migrateFlag := flag.String("migrate", "default", "For Migrate")

	flag.Parse()

	if cast.ToString(migrateFlag) == "fresh" {
		fmt.Println("Start Drop All Tables")

		for i := len(tables) - 1; i >= 0; i-- {
			fmt.Println("Dropping: " + tables[i].Name + " Table")

			err := dbConnection.Migrator().DropTable(tables[i].Model)

			if err != nil {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "04",
					"error": err.Error(),
				}).Error("failed to drop table")

				return
			}

			fmt.Println("Dropped: " + tables[i].Name + " Table")
		}

		fmt.Println("End Drop All Tables")
	}

	fmt.Println()

	fmt.Println("Start Migration")

	err = Migrate(dbConnection)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "05",
			"error": err.Error(),
		}).Error("failed to migrate")

		return
	}

	fmt.Println("End Migration")
}

func Migrate(db *gorm.DB) error {
	for _, t := range tables {
		fmt.Println("Migrating: " + t.Name + " Table")

		err := db.Migrator().AutoMigrate(t.Model)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   "internal.application.database.migration.main.Migrate.01",
				"error": err.Error(),
			}).Error("failed to create table")

			return err
		}

		fmt.Println("Migrated: " + t.Name + " Table")
	}

	return nil
}

package main

import (
	"flag"
	"fmt"

	"github.com/MrAndreID/goapi/configs"
	"github.com/MrAndreID/goapi/databases"
	"github.com/MrAndreID/goapi/databases/models"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

var tables map[string]interface{} = map[string]interface{}{
	"users":  &models.User{},
	"emails": &models.Email{},
}

func main() {
	var tag string = "Databases.Migrations.Main.Main."

	cfg, err := configs.New(true)

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
		dbConnection, err = databases.New(&databases.Database{
			Connection: cfg.DatabaseConnection,
			Host:       cfg.DatabaseHost,
			Port:       cfg.DatabasePort,
			Username:   cfg.DatabaseUsername,
			Password:   cfg.DatabasePassword,
			Name:       cfg.DatabaseName,
			SSLMode:    cfg.DatabaseSSLMode,
			ParseTime:  cfg.DatabaseParseTime,
			Charset:    cfg.DatabaseCharset,
			Timezone:   cfg.DatabaseTimezone,
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

		existingTables, err := dbConnection.Migrator().GetTables()

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "04",
				"error": err.Error(),
			}).Error("failed to get tables from database")

			return
		}

		for _, v := range existingTables {
			fmt.Println("Dropping: " + v + " Table")

			err := dbConnection.Migrator().DropTable(v)

			if err != nil {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "05",
					"error": err.Error(),
				}).Error("failed to drop table")

				return
			}

			fmt.Println("Dropped: " + v + " Table")
		}

		fmt.Println("End Drop All Tables")
	}

	fmt.Println()

	fmt.Println("Start Migration")

	err = Migrate(dbConnection)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "06",
			"error": err.Error(),
		}).Error("failed to migrate")

		return
	}

	fmt.Println("End Migration")
}

func Migrate(db *gorm.DB) error {
	for i, v := range tables {
		fmt.Println("Migrating: " + i + " Table")

		err := db.Migrator().AutoMigrate(v)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   "Databases.Migrations.Main.Migrate.01",
				"error": err.Error(),
			}).Error("failed to create table")

			return err
		}

		fmt.Println("Migrated: " + i + " Table")
	}

	return nil
}

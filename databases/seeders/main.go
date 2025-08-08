package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/MrAndreID/goapi/configs"
	"github.com/MrAndreID/goapi/databases"
	"github.com/MrAndreID/goapi/databases/models"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

var seeder map[string]map[string]interface{} = map[string]map[string]interface{}{
	"user1": {
		"model": &models.User{},
		"data": &models.User{
			ID:        "09123ae8-cce2-4d40-aac1-ae1b3c51cc77",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      "Andrea Adam",
		},
	},
	"user2": {
		"model": &models.User{},
		"data": &models.User{
			ID:        "7f5abfff-fae9-4c0d-8433-50f650583dac",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      "Zelda Skyward",
		},
	},
	"email1": {
		"model": &models.Email{},
		"data": &models.Email{
			ID:        "092fa1d6-aea8-4a0d-86d1-1c242d0f8ce5",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID:    "09123ae8-cce2-4d40-aac1-ae1b3c51cc77",
			Email:     "mrandreid.business@gmail.com",
		},
	},
	"email2": {
		"model": &models.Email{},
		"data": &models.Email{
			ID:        "902872a1-3c73-4fc5-8b9a-269203209d68",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID:    "09123ae8-cce2-4d40-aac1-ae1b3c51cc77",
			Email:     "andrea.adam.306147@brilian.bri.co.id",
		},
	},
	"email3": {
		"model": &models.Email{},
		"data": &models.Email{
			ID:        "61e5efb6-5da0-470f-a3ee-1109a2ea590e",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID:    "7f5abfff-fae9-4c0d-8433-50f650583dac",
			Email:     "zelda.skyward@email.com",
		},
	},
}

func main() {
	var tag string = "Databases.Seeders.Main.Main."

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

	fmt.Println("Start Seeder")

	seedFlag := flag.String("seed", "default", "For Seed")

	flag.Parse()

	if cast.ToString(seedFlag) == "default" {
		fmt.Println("Start Seed")

		for i, v := range seeder {
			fmt.Println("Seeding: " + i + " Data")

			for key, data := range v {
				if key == "model" {
					if !dbConnection.Migrator().HasTable(data) {
						logrus.WithFields(logrus.Fields{
							"tag":   tag + "04",
							"error": "Failed to Initiate Table",
						}).Error("failed to initiate table")

						return
					}
				}

				if key == "data" {
					result := dbConnection.Create(data)

					if result.Error != nil {
						logrus.WithFields(logrus.Fields{
							"tag":   tag + "05",
							"error": result.Error.Error(),
						}).Error("failed to create data")

						return
					}

					if result.RowsAffected == 0 {
						logrus.WithFields(logrus.Fields{
							"tag":   tag + "06",
							"error": "Failed to Create Data",
						}).Error("failed to create data")

						return
					}
				}
			}

			fmt.Println("Seeded: " + i + " Data")
		}

		fmt.Println("End Seed")
	}

	fmt.Println("End Seeder")
}

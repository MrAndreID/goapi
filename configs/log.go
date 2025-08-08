package configs

import (
	"io"
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

func NewBodyDumpLog() error {
	var tag string = "Configs.Log.NewBodyDumpLog."

	dir, err := os.Getwd()

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.Error(),
		}).Error("failed to get root path")

		return err
	}

	logfile, err := rotatelogs.New(
		dir+"/storages/logs/%Y%m%d.log",
		rotatelogs.WithLinkName(dir+"/storages/logs/master.log"),
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithMaxAge(-1),
		rotatelogs.WithRotationCount(365),
	)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to create a new rotate log")

		return err
	}

	logrus.SetFormatter(&logrus.JSONFormatter{DisableHTMLEscape: true})
	logrus.SetOutput(io.MultiWriter(os.Stdout, logfile))
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetReportCaller(true)

	return nil
}

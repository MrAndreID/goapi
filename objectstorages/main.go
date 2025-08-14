package objectstorages

import (
	"context"
	"errors"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

type ObjectStorage struct {
	Connection string
	Host       string
	Port       string
	Username   string
	Password   string
	SSL        bool
}

type ObjectStorageConnection struct {
	Minio *minio.Client
}

func New(objectStorage *ObjectStorage) (*ObjectStorageConnection, error) {
	var (
		objectStorageData *ObjectStorageConnection
		err               error
	)

	switch objectStorage.Connection {
	case "minio":
		objectStorageData, err = objectStorage.Minio()
	default:
		err = errors.New("Object Storage Connection Not Found")
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   "Object-Storages.Main.New.01",
			"error": err.Error(),
		}).Error("failed to connect object storage")

		return nil, err
	}

	return objectStorageData, nil
}

func (objectStorage *ObjectStorage) Minio() (*ObjectStorageConnection, error) {
	var (
		tag       string = "Object-Storages.Main.Minio."
		keyBucket string = "ping"
	)

	minioClient, err := minio.New(objectStorage.Host+":"+objectStorage.Port, &minio.Options{
		Creds:  credentials.NewStaticV4(objectStorage.Username, objectStorage.Password, ""),
		Secure: objectStorage.SSL,
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.Error(),
		}).Error("failed to connect minio")

		return nil, err
	}

	_, err = minioClient.BucketExists(context.Background(), keyBucket)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to connect minio")

		return nil, err
	}

	return &ObjectStorageConnection{
		Minio: minioClient,
	}, nil
}

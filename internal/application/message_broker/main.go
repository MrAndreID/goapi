package messagebroker

import (
	"context"
	"errors"
	"net/url"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type MessageBroker struct {
	Connection string
	Host       string
	Port       string
	Username   string
	Password   string
	Name       string
	Partition  int
}

type MessageBrokerConnection struct {
	Name     string
	RabbitMQ *RabbitMQConnection
	Kafka    *kafka.Conn
}

type RabbitMQConnection struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

func New(messageBroker *MessageBroker) (*MessageBrokerConnection, error) {
	var (
		messageBrokerConnection *MessageBrokerConnection
		err                     error
	)

	switch messageBroker.Connection {
	case "rabbitmq":
		messageBrokerConnection, err = messageBroker.RabbitMQ()
	case "kafka":
		messageBrokerConnection, err = messageBroker.Kafka()
	default:
		err = errors.New("Message Broker Connection Not Found")
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   "internal.application.messagebroker.main.New.01",
			"error": err.Error(),
		}).Error("failed to connect message broker")

		return nil, err
	}

	messageBrokerConnection.Name = messageBroker.Connection

	return messageBrokerConnection, nil
}

func (messageBroker *MessageBroker) RabbitMQ() (*MessageBrokerConnection, error) {
	var tag string = "internal.application.messagebroker.main.RabbitMQ."

	amqpURL := url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(messageBroker.Username, messageBroker.Password),
		Host:   messageBroker.Host + ":" + messageBroker.Port,
		Path:   "/",
	}

	rabbitMQConnection, err := amqp.Dial(amqpURL.String())

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.Error(),
		}).Error("failed to connect rabbitmq")

		return nil, err
	}

	rabbitMQChannel, err := rabbitMQConnection.Channel()

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to connect rabbitmq")

		return nil, err
	}

	return &MessageBrokerConnection{
		RabbitMQ: &RabbitMQConnection{
			Connection: rabbitMQConnection,
			Channel:    rabbitMQChannel,
		},
	}, nil
}

func (messageBroker *MessageBroker) Kafka() (*MessageBrokerConnection, error) {
	kafkaConnection, err := kafka.DialLeader(context.Background(), "tcp", messageBroker.Host+":"+messageBroker.Port, messageBroker.Name, messageBroker.Partition)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   "internal.application.messagebroker.main.Kafka.01",
			"error": err.Error(),
		}).Error("failed to connect kafka")

		return nil, err
	}

	return &MessageBrokerConnection{
		Kafka: kafkaConnection,
	}, nil
}

func (messageBrokerConnection *MessageBrokerConnection) Close() {
	var tag string = "internal.application.messagebroker.main.Close."

	switch messageBrokerConnection.Name {
	case "rabbitmq":
		if messageBrokerConnection.RabbitMQ == nil {
			return
		}

		if messageBrokerConnection.RabbitMQ.Channel != nil {
			if err := messageBrokerConnection.RabbitMQ.Channel.Close(); err != nil {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "01",
					"error": err.Error(),
				}).Error("failed to close rabbitmq channel")
			}
		}

		if messageBrokerConnection.RabbitMQ.Connection != nil {
			if err := messageBrokerConnection.RabbitMQ.Connection.Close(); err != nil {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "02",
					"error": err.Error(),
				}).Error("failed to close rabbitmq connection")
			}
		}
	case "kafka":
		if messageBrokerConnection.Kafka != nil {
			if err := messageBrokerConnection.Kafka.Close(); err != nil {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "03",
					"error": err.Error(),
				}).Error("failed to close kafka connection")
			}
		}
	}
}

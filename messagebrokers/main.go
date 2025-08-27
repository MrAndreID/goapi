package messagebrokers

import (
	"context"
	"errors"

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
			"tag":   "Message-Brokers.Main.New.01",
			"error": err.Error(),
		}).Error("failed to connect message broker")

		return nil, err
	}

	messageBrokerConnection.Name = messageBroker.Connection

	return messageBrokerConnection, nil
}

func (messageBroker *MessageBroker) RabbitMQ() (*MessageBrokerConnection, error) {
	var tag string = "Message-Brokers.Main.RabbitMQ."

	rabbitMQConnection, err := amqp.Dial("amqp://" + messageBroker.Username + ":" + messageBroker.Password + "@" + messageBroker.Host + ":" + messageBroker.Port + "/")

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
			"tag":   "Message-Brokers.Main.Kafka.01",
			"error": err.Error(),
		}).Error("failed to connect kafka")

		return nil, err
	}

	return &MessageBrokerConnection{
		Kafka: kafkaConnection,
	}, nil
}

func (messageBrokerConnection *MessageBrokerConnection) Close() {
	var err error

	switch messageBrokerConnection.Name {
	case "rabbitmq":
		messageBrokerConnection.RabbitMQ.Connection.Close()
		messageBrokerConnection.RabbitMQ.Channel.Close()
	case "kafka":
		messageBrokerConnection.Kafka.Close()
	default:
		err = errors.New("Message Broker Connection Not Found")
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   "Message-Brokers.Main.Close.01",
			"error": err.Error(),
		}).Error("failed to close connection (message broker)")
	}
}

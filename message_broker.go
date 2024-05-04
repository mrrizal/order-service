package main

import "github.com/streadway/amqp"

func declareQueue() error {
	conn, err := amqp.Dial(BrokerURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		QueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	return nil
}

func initRabbitMQ() (*amqp.Connection, error) {
	conn, err := amqp.Dial(BrokerURL)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func closeRabbitMQ(conn *amqp.Connection) error {
	if conn != nil {
		return conn.Close()
	}
	return nil
}

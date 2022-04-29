package rabbitmq

import (
	"context"
	"fmt"
	"log"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

const (
	ExchangeAuth = "auth"
	ExchangeGate = "gate"
	ExchangeChat = "chat"
)

const (
	RouteWrite = "write"

	RouteLogin    = "login"
	RouteRegister = "register"

	RouteSendMsg    = "sendMsg"
	RouteReceiveMsg = "receiveMsg"
)

// AmqpConn is our real implementation, encapsulates a pointer to an amqp.Connection
type AmqpConn struct {
	Conn *amqp.Connection
}

// connectToBroker connects to an AMQP broker using the supplied connectionString.
func (c *AmqpConn) connectToBroker(ctx context.Context, connectionString string) {
	if connectionString == "" {
		panic("Cannot initialize connection to broker, connectionString not set. Have you initialized?")
	}

	var err error
	c.Conn, err = amqp.Dial(fmt.Sprintf("%s/", connectionString))
	if err != nil {
		panic("Failed to connect to AMQP compatible broker at: " + connectionString)
	}
}

// PublishToTopic .
func (c *AmqpConn) PublishToTopic(ctx context.Context, exchangeName, routingKey string, body []byte) error {
	ch, err := c.Conn.Channel()
	if err != nil {
		return errors.Wrap(err, "Failed to open a channel")
	}
	defer func() {
		err := ch.Close()
		if err != nil {
			log.Println("close the channel err is:", err.Error())
		}
	}()

	err = ch.ExchangeDeclare(
		exchangeName, // exchange name
		"topic",      // exchange kind
		true,         // durable
		false,        // auto delete
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "failed to ExchangeDeclare")
	}
	err = ch.Publish(
		exchangeName, // exchange
		routingKey,   // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		return errors.Wrap(err, "failed to register an consumer")
	}

	// log.Printf("A message was sent to exchangeName: %v routingKey: %v data: %v\n", exchangeName, routingKey, string(body))

	return nil
}

func (c *AmqpConn) GetPublishChannel(ctx context.Context, kind, exchangeName string) (*AMQPChannel, error) {
	ch, err := c.Conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to open a channel")
	}
	err = ch.ExchangeDeclare(
		exchangeName, // exchange name
		kind,         // exchange kind
		true,         // durable
		false,        // auto delete
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ExchangeDeclare")
	}
	return &AMQPChannel{
		channel: ch,
	}, nil
}

// SubscribeFromTopic .
func (c *AmqpConn) SubscribeFromTopic(ctx context.Context, exchangeName string, routerKeys []string, queueName string,
	handlerFunc func(amqp.Delivery)) error {
	ch, err := c.Conn.Channel()
	if err != nil {
		return errors.Wrap(err, "Failed to open a channel")
	}

	defer func() {
		err := ch.Close()
		if err != nil {
			log.Println("close the channel err is:", err.Error())
		}
	}()

	err = ch.ExchangeDeclare(
		exchangeName, // exchange name
		"topic",      // exchange kind
		true,         // durable
		false,        // auto delete
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to open a ExchangeDeclare")
	}

	q, err := ch.QueueDeclare(
		queueName,
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return errors.Wrap(err, "Failed to open a QueueDeclare")
	}

	if len(routerKeys) == 0 {
		routerKeys = append(routerKeys, "")
	}

	for _, key := range routerKeys {
		err = ch.QueueBind(
			q.Name,
			key,
			exchangeName,
			false,
			nil,
		)
		if err != nil {
			return errors.Wrap(err, "Failed to QueueBind")
		}
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false, // Auto Ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to Consume msgs")
	}

	consumeLoop(ctx, msgs, handlerFunc)
	return nil
}

// Close closes the connection to the AMQP-broker, if available.
func (c *AmqpConn) Close() {
	if c.Conn != nil {
		log.Println("Closing connection to AMQP broker")
		_ = c.Conn.Close()
	}
}

func consumeLoop(ctx context.Context, deliveries <-chan amqp.Delivery, handlerFunc func(d amqp.Delivery)) {
	for d := range deliveries {
		// Invoke the handlerFunc func we passed as parameter.
		handlerFunc(d)
	}
}

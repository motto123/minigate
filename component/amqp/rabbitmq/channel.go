package rabbitmq

import (
	"context"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type AMQPChannel struct {
	channel *amqp.Channel
}

func (ch *AMQPChannel) SendMessage(ctx context.Context, exchangeName, routerKey string, body []byte) error {
	err := ch.channel.Publish(
		exchangeName, // exchange
		routerKey,    // routing key
		false,
		false,
		amqp.Publishing{
			Headers:         map[string]interface{}{},
			DeliveryMode:    amqp.Persistent,
			MessageId:       ctx.Value("message_id").(string),
			Body:            body,
		})

	if err != nil {
		return errors.Wrap(err, "failed to register an consumer")
	}

	// log.Printf("A message was sent to exchangeName: %v routingKey: %v data: %v\n", exchangeName, routingKey, string(body))

	return nil
}

func (ch *AMQPChannel) Close(ctx context.Context) error {
	return ch.channel.Close()
}

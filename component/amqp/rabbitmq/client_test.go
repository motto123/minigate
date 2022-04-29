package rabbitmq

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/spf13/cast"
	"github.com/streadway/amqp"
)

var mqConn *AmqpConn

func setup() {
	var err error
	mqConn, err = NewConn(context.Background(), &RabbitInitConfig{
		UserName: "admin",
		Password: "pass.123",
		Host:     "192.168.11.13",
		Port:     5672,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func teardown() {
	mqConn.Close()
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func TestMessageProduce(t *testing.T) {
	i := 1
	for {
		if err := mqConn.PublishToTopic(context.Background(), "test", "1", []byte("test:"+cast.ToString(i))); err != nil {
			t.Error("test", "PublishToTopic error:", err)
		}
		i++
		time.Sleep(1 * time.Second)
	}
}

func TestSendMessageByChannel(t *testing.T) {
	i := 1
	ch, err := mqConn.GetPublishChannel(context.Background(), "topic", "test")
	if err != nil {
		t.Error("test", "GetPublishChannel error:", err)
	}
	for {
		if err := ch.SendMessage(context.Background(), "test", "1", []byte("test:"+cast.ToString(i))); err != nil {
			t.Error("test", "PublishToTopic error:", err)
		}
		i++
	}
}

func TestMessageConsumer(t *testing.T) {
	ctx := context.Background()
	go func() {
		err := mqConn.SubscribeFromTopic(ctx, "test", []string{"1"}, "testQ1", onTestMessage1)
		if err != nil {
			log.Fatalf("receive message error: %s \n", err.Error())
		}
	}()
	err := mqConn.SubscribeFromTopic(ctx, "test", []string{"1"}, "testQ1", onTestMessage2)
	if err != nil {
		log.Fatalf("receive message error: %s \n", err.Error())
	}
}

func onTestMessage1(delivery amqp.Delivery) {
	log.Printf("the testQ1 amqp.Delivery body is: %s \n", string(delivery.Body))
	err := delivery.Ack(false)
	if err != nil {
		log.Fatal("ack error:", err)
	}
}

func onTestMessage2(delivery amqp.Delivery) {
	log.Printf("the testQ2 amqp.Delivery body is: %s \n", string(delivery.Body))
	err := delivery.Ack(false)
	if err != nil {
		log.Fatal("ack error:", err)
	}
}

package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/spf13/viper"
)

type RabbitInitConfig struct {
	UserName string
	Password string
	Host     string
	Port     int
}

// NewConn method
func NewConn(ctx context.Context, initConfig *RabbitInitConfig) (*AmqpConn, error) {
	if initConfig == nil {
		initConfig = &RabbitInitConfig{
			UserName: viper.GetString("amqp.rabbitmq.username"),
			Password: viper.GetString("amqp.rabbitmq.password"),
			Host:     viper.GetString("amqp.rabbitmq.host"),
			Port:     viper.GetInt("amqp.rabbitmq.port"),
		}
	}
	var (
		client *AmqpConn
		once   sync.Once
	)
	once.Do(func() {
		client = new(AmqpConn)

		// amqp://用户名:密码@地址:端口号
		connectStr := fmt.Sprintf("amqp://%s:%s@%s:%d",
			initConfig.UserName,
			initConfig.Password,
			initConfig.Host,
			initConfig.Port,
		)
		client.connectToBroker(ctx, connectStr)

		log.Println("rabbitmq connect successfully")
	})
	return client, nil
}

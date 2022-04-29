module com.minigame.component

go 1.16

replace com.minigame.proto => ../proto

replace com.minigame.utils => ../utils

require (
	com.minigame.proto v0.0.0-00010101000000-000000000000
	com.minigame.utils v0.0.0-00010101000000-000000000000
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-sql-driver/mysql v1.6.0
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/lestrrat-go/strftime v1.0.5 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.18.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cast v1.4.1
	github.com/spf13/viper v1.10.1
	github.com/streadway/amqp v1.0.0
	github.com/stretchr/testify v1.7.1
	google.golang.org/protobuf v1.28.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

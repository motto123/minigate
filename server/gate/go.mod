module com.minigame.server.gate

go 1.16

replace (
	com.minigame.component => ../../component
	com.minigame.proto => ../../proto
	com.minigame.utils => ../../utils
)

require (
	com.minigame.component v0.0.0-00010101000000-000000000000
	com.minigame.proto v0.0.0-00010101000000-000000000000
	com.minigame.utils v0.0.0-00010101000000-000000000000
	github.com/bwmarrin/snowflake v0.3.0
	github.com/gorilla/websocket v1.5.0
	github.com/pkg/errors v0.9.1
	github.com/streadway/amqp v1.0.0
	google.golang.org/grpc v1.46.0
	google.golang.org/protobuf v1.28.0
)

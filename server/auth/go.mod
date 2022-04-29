module com.minigame.auth

go 1.16

replace com.minigame.component => ../../component

replace com.minigame.proto => ../../proto

replace com.minigame.utils => ../../utils

require (
	com.minigame.component v0.0.0-00010101000000-000000000000
	com.minigame.proto v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	google.golang.org/grpc v1.46.0
	google.golang.org/protobuf v1.28.0
)

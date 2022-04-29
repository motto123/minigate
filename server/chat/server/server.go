package server

import (
	"com.minigame.component/amqp/rabbitmq"
	"com.minigame.component/base"
	"com.minigame.component/log"
	"com.minigame.proto/gate"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	tag = "ChatServer"
)

var (
	Srv = new(ChatServer) // 创建服务实例
)

type ChatServer struct {
	base.BusinessServer
	conf       *config
	gateSrvCli gate.GateSrvClient
}

func (s *ChatServer) Init(fw *base.Framework) error {
	s.SetTag(tag)
	err := s.BusinessServer.Init(fw)
	if err != nil {
		return err
	}
	err = s.loadConf()
	if err != nil {
		return err
	}
	if s.conf.Server.RunningModel == "debug" {
		s.SetIsDebug(true)
	}
	err = s.initGrpcClient()
	if err != nil {
		return err
	}

	s.registerRouter()

	return nil
}

func (s *ChatServer) loadConf() (err error) {
	s.conf = new(config)
	s.Vipper.SetConfigName("server")
	err = s.Vipper.ReadInConfig()
	err = errors.WithStack(err)
	if err != nil {
		return
	}
	err = s.Vipper.Unmarshal(s.conf)
	err = errors.WithStack(err)
	return
}

func (s *ChatServer) initGrpcClient() (err error) {
	addr := "127.0.0.1:6701"
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		err = errors.WithStack(err)
		log.Fatalf(tag, "grpc.Dial failed, err: %+v", err)
	}
	s.gateSrvCli = gate.NewGateSrvClient(conn)
	return
}

func (s *ChatServer) registerRouter() {
	s.RegisterRouter(rabbitmq.RouteSendMsg, sendMsgHandler)
	s.RegisterRouter(rabbitmq.RouteReceiveMsg, receiveMsgHandler)

	s.ListenRouter(rabbitmq.ExchangeChat)
}

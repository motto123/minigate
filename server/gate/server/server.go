package server

import (
	"com.minigame.component/amqp/rabbitmq"
	"com.minigame.component/base"
	"com.minigame.component/codec/message"
	"com.minigame.component/codec/packet"
	"com.minigame.component/codec/router"
	"com.minigame.component/log"
	"com.minigame.proto/common"
	"com.minigame.proto/gate"
	"com.minigame.server.gate/session"
	utils "com.minigame.utils"
	"context"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"net"
	"net/http"
	"strings"
	"time"
)

var tag = "GateServer"

var Srv = new(GateServer)

// GateServer 可以部署多个gate server,提高吞吐量,gate拥有管理 session,
// 通过路由gate可以将用户request自动传给business server中的handlerFunc
// 程序员只关心这个initRouter就行了,只要在 initRouter 中注册路由,business server中注册相同的路由，就可以实现业务功能
// 为gate grpc 增加更多hander,可以让business server和gate sever有跟多的互动
type GateServer struct {
	base.BaseServer
	conf      *config
	IdGeneral *snowflake.Node
	//注销session的channel
	unregister chan uint64
	//接受session.read的数据channel
	send chan *session.PostMsg
	//路由
	router   *router.Router
	sessionM *session.SyncClientMap
	isDebug  bool
}

func (s *GateServer) Init(*base.Framework) error {
	s.SetTag(tag)
	err := s.BaseServer.Init(nil)
	if err != nil {
		return err
	}
	err = s.loadConf()
	if err != nil {
		return err
	}

	if s.conf.Server.RunningModel == "debug" {
		s.isDebug = true
	}

	err = s.initSnowflakeId()
	if err != nil {
		return err
	}

	s.sessionM = session.NewSyncSessionMap()
	s.unregister = make(chan uint64, 0)
	// TODO: 此处可优化,以免channel造成瓶颈
	s.send = make(chan *session.PostMsg, 1024)
	s.router = router.NewRouter()

	err = s.registerGrpc()
	if err != nil {
		return err
	}
	err = s.registerTcp()
	if err != nil {
		return err
	}
	err = s.registerWs()
	if err != nil {
		return err
	}

	s.unregisterRecv()
	s.sendRecv()
	s.writeRecv()
	s.initRouter()
	return nil
}

func (s *GateServer) loadConf() (err error) {
	s.conf = new(config)
	s.Vipper.SetConfigName("server")
	err = s.Vipper.ReadInConfig()
	err = errors.WithStack(err)
	if err != nil {
		return
	}
	err = s.Vipper.Unmarshal(s.conf)
	err = errors.WithStack(err)
	log.Infof(tag, "net conf: %+v", s.conf)
	return
}

func (s *GateServer) initSnowflakeId() error {
	var err error
	s.IdGeneral, err = snowflake.NewNode(s.conf.Server.Node)
	err = errors.WithStack(err)
	if err != nil {
		return err
	}
	return nil
}

func (s *GateServer) registerGrpc() (err error) {
	// 用于businessServer和gate互动使用
	// 比如操作session
	address := fmt.Sprintf("%s:%s", s.conf.Grpc.Ip, s.conf.Grpc.Port)
	listen, err := net.Listen("tcp", address)
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf(tag, "net.Listen failed, err: %+v", err)
		return
	}

	gs := grpc.NewServer(
		grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			defer func() {
				if e := recover(); e != nil {
					//log.Errorf(tag, "grpc panic:%s", string(debug.Stack()))
					s.CatchError(e)
				}
			}()
			name := info.FullMethod[strings.LastIndex(info.FullMethod, "/")+1:]
			log.Debugf(tag, "[rpc] recv %s, req: {%+v}", name, req)
			resp, err = handler(ctx, req)
			log.Debugf(tag, "[rpc] ret %s, resp: {%+v}", name, resp)
			return
		}),
	)
	gate.RegisterGateSrvServer(gs, new(GateGrpcSrv))
	go func() {
		err = gs.Serve(listen)
		if err != nil {
			err = errors.WithStack(err)
			log.Errorf(tag, "gs.Serve failed, err: %+v", err)
			return
		}
	}()
	log.Infof(tag, "grpc listen on %s", address)
	return
}

func (s *GateServer) registerTcp() (err error) {
	// 创建 listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", s.conf.Tcp.Port))
	err = errors.WithStack(err)
	if err != nil {
		//log.Fatalf(tag, "listening failed, err: %s", err.Error())
		return
	}
	log.Infof(tag, "tcp listen :%s", s.conf.Tcp.Port)

	// 监听并接受来自客户端的连接
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatalf(tag, "accepting failed, err: %s", err.Error())
			}
			id := s.IdGeneral.Generate().Int64()

			newSession := session.NewTcpSession(uint64(id), conn, s.unregister, s.send, s.router,
				s.conf.Debug.IsCheckHeartbeat, s.conf.Debug.IsRecordHeartbeatLog)
			newSession.Init()
			newSession.ReadPump()
			s.sessionM.Store(newSession.Id, newSession)
		}
	}()
	return nil

}

func (s *GateServer) registerWs() (err error) {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("this is home"))
	})
	http.HandleFunc("/web", func(writer http.ResponseWriter, request *http.Request) {
		s.serveWs(writer, request)
	})
	go func() {
		err = http.ListenAndServe(fmt.Sprintf(":%s", s.conf.Websocket.Port), nil)
		err = errors.WithStack(err)
		if err != nil {
			return
		}
		log.Infof(tag, "ws listen :%s", s.conf.Websocket.Port)
	}()
	return nil
}

func (s *GateServer) serveWs(w http.ResponseWriter, r *http.Request) {
	fmt.Println("websocket.Upgrader")
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true }}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf(tag, "upgrader.Upgrade failed, err: %s", err)
		return
	}
	id := uint64(s.IdGeneral.Generate().Int64())

	//newSession := session.NewWsSession(id, conn, s.hub.Register, s.hub.Unregister, s.hub.send)
	newSession := session.NewWsSession(id, conn, s.unregister, s.send, s.conf.Debug.IsCheckHeartbeat,
		s.conf.Debug.IsRecordHeartbeatLog)
	newSession.Init()
	newSession.ReadPump()
	s.sessionM.Store(id, newSession)
}

// unregisterRecv 统计处理session注销
func (s *GateServer) unregisterRecv() {
	go func() {
		for {
			sessionId := <-s.unregister
			session2 := s.sessionM.Load(sessionId)
			if session2 == nil {
				continue
			}
			err := session2.Close()
			if err != nil {
				log.Errorf(tag, "session2.Close failed, err: %+v", err)
				continue
			}
			s.sessionM.Delete(sessionId)
			log.Infof(tag, "[gate] unregister sid: %d, uid: %d", session2.GetId(), session2.GetUid())
		}
	}()

}

// sendRecv 所有session.read的数据都要通过mq传给businessServer中的具体handlerFunc处理
func (s *GateServer) sendRecv() {
	go func() {
		for {
			postMsg := <-s.send
			session2 := s.sessionM.Load(postMsg.SessionId)
			if session2 == nil {
				return
			}
			for _, byts := range postMsg.MsgBytesList {
				msg := message.NewMessageWithRouter(s.router)
				err := msg.Decode(byts)
				if err != nil {
					log.Errorf(tag, "msg.Decode failed, err: %+v", err)
					continue
				}
				exchangeName, ok := s.router.GetExchangeName(msg.Route)
				if !ok {
					_, err = log.ErrorfAndRetErr(tag, "illegal route %s", msg.Route)
					s.generalResponseErr(session2, msg, err)
					continue
				}

				log.Debugf(tag, "%c[1;0;36m [pub mq],msgId: %d, from sid: %d,  msg: {type:%s route:%s exchange:%s}, %c[0m",
					0x1B, msg.Id, postMsg.SessionId, msg.Type, msg.Route, exchangeName, 0x1B)

				if !(msg.Type == message.Request || msg.Type == message.Notify) {
					_, err = log.ErrorfAndRetErr(tag, "illegal message type %s from session id", msg.Type, msg.Id, postMsg.SessionId)
					s.generalResponseErr(session2, msg, err)
					continue
				}

				info := &gate.SendInfo{
					Sid:  postMsg.SessionId,
					Body: byts,
				}
				body, err := proto.Marshal(info)
				err = errors.WithStack(err)
				if err != nil {
					log.Errorf(tag, "proto.Marshal failed, err: %+v", err)
					s.generalResponseErr(session2, msg, err)
					continue
				}

				err = s.MQConn.PublishToTopic(s.Context, exchangeName, msg.Route, body)
				if err != nil {
					log.Errorf(tag, "MQConn.PublishToTopic failed, err: %+v", err)
					s.generalResponseErr(session2, msg, err)
					continue
				}
			}
		}
	}()

}

// writeRecv businessServer中的handlerFunc处理结束后给session.write的数据
func (s *GateServer) writeRecv() {
	go func() {
		handlerFunc := func(delivery amqp.Delivery) {
			defer func() {
				err := delivery.Ack(true)
				if err != nil {
					log.Errorf(tag, "delivery.Ack failed, err: %+v", err)
				}
			}()
			if len(delivery.Body) == 0 {
				log.Errorf(tag, "delivery.Body len is 0")
				return
			}

			ri := new(gate.ReceiveInfo)
			err := proto.Unmarshal(delivery.Body, ri)
			if err != nil {
				err = errors.WithStack(err)
				log.Errorf(tag, "proto.Unmarshal failed, err: %+v", err)
				return
			}

			log.Debugf(tag, "%c[1;0;36m [sub mq] msgId %d, write to ri: {recvSids: %v, recvUids: %v}, msg: {type:%s, dataLen:%d}%c[0m",
				0x1B, ri.MsgId, ri.ReceiverSids, ri.ReceiverUids, message.Type(ri.MsgType), len(ri.Body), 0x1B)
			for _, id := range ri.ReceiverSids {
				sessionInfo := s.sessionM.Load(id)
				if sessionInfo == nil {
					log.Infof(tag, "s.sessionM.Load failed, because sessionInfo id %d is nil", id)
					continue
				}

				if s.conf.Debug.IsCheckHeartbeat {
					err := sessionInfo.SetWriteDeadline(time.Now().Add(session.WriteLine))
					if err != nil {
						log.Errorf(tag, "sessionInfo.SetWriteDeadline failed, err: %+v", err)
						continue
					}
				}
				_, err = sessionInfo.Write(ri.Body)
				if err != nil {
					log.Errorf(tag, "sessionInfo.Write failed, err: %+v", err)
					continue
				}
			}

			for _, uid := range ri.ReceiverUids {
				session2 := s.sessionM.LoadWithUid(uid)
				if session2 == nil {
					log.Infof(tag, "s.sessionM.Load failed, because session2 id %d is nil", session2.GetId())
					continue
				}

				if s.conf.Debug.IsCheckHeartbeat {
					err := session2.SetWriteDeadline(time.Now().Add(session.WriteLine))
					if err != nil {
						log.Errorf(tag, "session2.SetWriteDeadline failed, err: %+v", err)
						continue
					}
				}
				_, err = session2.Write(ri.Body)
				if err != nil {
					log.Errorf(tag, "session2.Write failed, err: %+v", err)
					continue
				}
			}
		}

		err := s.MQConn.SubscribeFromTopic(s.Context, rabbitmq.ExchangeGate, []string{rabbitmq.RouteWrite}, "",
			handlerFunc)
		if err != nil {
			err = errors.WithStack(err)
			log.Errorf(tag, "MQConn.SubscribeFromTopic failed, err: %+v", err)
		}
	}()
}

// initRouter  注册路由，用于自动压缩路由，还可以通过路由找到处理这个路由的handlerFunc所在的businessServer
func (s *GateServer) initRouter() {
	//s.router.AddRouteKV(rabbitmq.ExchangeAuth, rabbitmq.RouteLogin, 1)
	//s.router.AddRouteKV(rabbitmq.ExchangeAuth, rabbitmq.RouteRegister, 2)

	s.router.AddRoute(rabbitmq.ExchangeAuth, rabbitmq.RouteLogin)
	s.router.AddRoute(rabbitmq.ExchangeAuth, rabbitmq.RouteRegister)

	s.router.AddRoute(rabbitmq.ExchangeChat, rabbitmq.RouteSendMsg)
	s.router.AddRoute(rabbitmq.ExchangeChat, rabbitmq.RouteReceiveMsg)

}

// generalResponseErr 生成通用的错误Response
func (s *GateServer) generalResponseErr(session2 session.Session, msg *message.Message, err error) {
	msg.ErasureDataInfoWithoutOther()
	re := &gate.ResponseErr{Code: &common.ResponseCode{Code: common.Code_InternalError}}
	if s.isDebug && err != nil {
		re.Code.Err = err.Error()
	}
	marshal, err := proto.Marshal(re)
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf(tag, "proto.Marshal failed, err: %+v", err)
		return
	}
	msg.Type = message.Response
	msg.Data = marshal
	msg.DataType = message.Protobuf
	msg.DataObjName = utils.GetStructName(re)
	msgBuf, err := msg.Encode()
	if err != nil {
		log.Errorf(tag, "msg.Encode failed, err: %+v", err)
		return
	}
	body, err := session2.GetDecoder().Encode(packet.Data, msgBuf)
	if err != nil {
		log.Errorf(tag, "session2.GetDecoder().Encode failed, err: %+v", err)
		return
	}
	_, err = session2.Write(body)
	if err != nil {
		log.Errorf(tag, "session2.Write failed, err: %+v", err)
		return
	}
	return
}

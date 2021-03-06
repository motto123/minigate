package base

import (
	"com.minigame.component/amqp/rabbitmq"
	"com.minigame.component/codec"
	"com.minigame.component/codec/message"
	"com.minigame.component/codec/packet"
	"com.minigame.component/log"
	"com.minigame.proto/common"
	"com.minigame.proto/gate"
	utils "com.minigame.utils"
	"fmt"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"google.golang.org/protobuf/proto"
)

// handlerFunc 业务代码执行函数
type handlerFunc func(s *BusinessServer, sid uint64, msg *message.Message,
	ri *gate.ReceiveInfo) (reqMsg, respMsg proto.Message)

// BusinessServer 业务服务使用的基础sever,提供mq subject的publish和router一套逻辑代码
type BusinessServer struct {
	BaseServer
	decoder *codec.Decoder
	// 路由和业务函数map
	handlerFuncM map[string]handlerFunc
	isDebug      bool
}

func (s *BusinessServer) Init(fw *Framework) error {
	err := s.BaseServer.Init(fw)
	if err != nil {
		return err
	}

	s.decoder = codec.NewDecoder()
	s.handlerFuncM = make(map[string]handlerFunc)
	return nil
}

func (s *BusinessServer) SetTag(str string) {
	tag = fmt.Sprintf("%s.(%s)", "BusinessServer", str)
}

func (s *BusinessServer) SetIsDebug(b bool) {
	s.isDebug = b
}

// ListenRouter 监听制定的exchange中的数据,routeKey对应gate中的route,
// 把数据交给指定的路由对应的handlerFunc
func (s *BusinessServer) ListenRouter(exchangeName string) {
	go func() {
		// 获取所有路由的名字
		var routerKeys []string
		for k, _ := range s.handlerFuncM {
			routerKeys = append(routerKeys, k)
		}
		if len(routerKeys) == 0 {
			log.Errorf(tag, "route keys is empty")
			return
		}

		// 监听这些路由的数据
		err := s.MQConn.SubscribeFromTopic(s.Context, exchangeName, routerKeys, "",
			func(delivery amqp.Delivery) {
				defer func() {
					err := delivery.Ack(true)
					err = errors.WithStack(err)
					if err != nil {
						log.Errorf(tag, "delivery.Ack failed, err: %+v", err)
						return
					}
				}()

				//fmt.Printf("delivery.Body: %v\n", delivery.Body)
				if len(delivery.Body) == 0 {
					//TODO: 完善错误提示
					//s.publishErrInfo(delivery, nil)
					return
				}
				si := new(gate.SendInfo)
				err := proto.Unmarshal(delivery.Body, si)
				err = errors.WithStack(err)
				if err != nil {
					log.Errorf(tag, "proto.Unmarshal failed, err: %+v", err)
					return
				}

				// 创建一个不包含router的message,这样可以不用解析传来的message中的路由,
				// message.Encode时不对路由做任何处理，保持原样
				// 为什么: 因为BusinessServer没有路由字典,不能对压缩路由进行解析,
				//  TODO: 下个版本解决
				msg := message.NewMessageAndUnchangedRoute()
				err = msg.Decode(si.Body)
				if err != nil {
					return
				}

				h, ok := s.handlerFuncM[delivery.RoutingKey]
				if !ok {
					log.Errorf(tag, "s.handlerFuncM[key] failed, illegal route key")
					return
				}
				// 执行handlerFnc,并打印日志
				ri, isPanic := s.wrapHandlerFunc(delivery.RoutingKey, si.Sid, msg, h)
				if isPanic {
					return
				}
				//允许路由handler执行完后,不返回数据
				if len(ri.ReceiverSids) == 0 && len(ri.ReceiverUids) == 0 {
					return
				}

				body, err := s.generalReceiveInfoBytes(packet.Data, msg, ri)
				if err != nil {
					return
				}

				err = s.MQConn.PublishToTopic(s.Context, rabbitmq.ExchangeGate, rabbitmq.RouteWrite, body)
				err = errors.WithStack(err)
				if err != nil {
					log.Errorf(tag, "MQConn.PublishToTopic failed, err: %+v", err)
					return
				}
			})
		log.Errorf(tag, "MQConn.SubscribeFromTopic failed, err: %+v", err)

		return
	}()
}

func (s *BusinessServer) PublishDataToWriteChanel(msg *message.Message, ri *gate.ReceiveInfo) error {
	body, err := s.generalReceiveInfoBytes(packet.Data, msg, ri)
	if err != nil {
		return err
	}

	err = s.MQConn.PublishToTopic(s.Context, rabbitmq.ExchangeGate, rabbitmq.RouteWrite, body)
	err = errors.WithStack(err)
	if err != nil {
		return err
	}
	return nil
}

func (s *BusinessServer) publishErrInfo(delivery amqp.Delivery, marshal []byte, sessionIds []uint64) {
	defer func() {
		err := delivery.Ack(true)
		err = errors.WithStack(err)
		if err != nil {
			log.Errorf(tag, "delivery.Ack failed, err: %+v", err)
			return
		}
	}()

	//body, err := s.generalReceiveInfoBytes(packet.Data, marshal, sessionIds)
	//if err != nil {
	//	return
	//}
	//err = s.MQConn.PublishToTopic(s.Context, rabbitmq.ExchangeGate, rabbitmq.RouteWrite, body)
	//err = errors.WithStack(err)
	//if err != nil {
	//	log.Errorf(tag, "MQConn.PublishToTopic failed, err: %+v", err)
	//	return
	//}
}

// RegisterRouter 注册路由和handlerFunc
func (s *BusinessServer) RegisterRouter(routeKey string, handleFunc handlerFunc) {
	s.handlerFuncM[routeKey] = handleFunc
}

func (s *BusinessServer) marshalForProto(routeKey string, sessionId uint64, msg *message.Message,
	handlerFunc handlerFunc) (ri *gate.ReceiveInfo, reqMsg, respMsg proto.Message) {
	ri = new(gate.ReceiveInfo)
	reqMsg, respMsg = handlerFunc(s, sessionId, msg, ri)

	if len(ri.ReceiverSids) == 0 && len(ri.ReceiverUids) == 0 {
		return
	}
	if respMsg != nil { //允许msg.Data中不携带数据
		marshal, err := proto.Marshal(respMsg)
		if err != nil {
			err = errors.WithStack(err)
			log.Errorf(tag, "proto.Marshal failed, err: %+v", err)
			return
		}
		msg.ErasureDataInfoWithoutOther()
		msg.DataType = message.Protobuf
		msg.Data = marshal
		msg.DataObjName = utils.GetStructName(respMsg)
	}
	return
}

func (s *BusinessServer) generalReceiveInfoBytes(pkgType packet.Type, msg *message.Message, ri *gate.ReceiveInfo) (body []byte, err error) {
	msgBuf, err := msg.Encode()
	if err != nil {
		log.Errorf(tag, "msg.Encode failed, err: %+v", err)
		return
	}

	pkgBuf, err := s.decoder.Encode(pkgType, msgBuf)
	if err != nil {
		log.Errorf(tag, "decoder.Encode failed, err: %+v", err)
		return
	}

	ri.MsgId = msg.Id
	ri.MsgType = int32(msg.Type)
	ri.Body = pkgBuf

	body, err = proto.Marshal(ri)
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf(tag, "proto.Marshal failed, err: %+v", err)
		return
	}
	return
}

// DefaultErrProcess 测试模式返回err便于调试,正式模式只返回 code前端用proto转化成具体的错误信息,
// 这样做可以缩小数据包的大小
func (s *BusinessServer) DefaultErrProcess(rc *common.ResponseCode, err error) {
	//rc.Err = rc.Code.String()
	if !s.isDebug || err == nil {
		return
	}
	rc.Err = err.Error()
}

func (s *BusinessServer) wrapHandlerFunc(routeKey string, sessionId uint64, msg *message.Message,
	handlerFunc handlerFunc) (ri *gate.ReceiveInfo, isPanic bool) {
	if handlerFunc == nil {
		return
	}
	fn := utils.GetFunctionName(handlerFunc)
	defer func() {
		if e := recover(); e != nil {
			s.CatchError(fmt.Sprintf("panic: execute func %s failed, err: %s", fn, e))
			isPanic = true
		}
	}()

	msgTypeBefore := msg.Type
	ri, reqMsg, respMsg := s.marshalForProto(routeKey, sessionId, msg, handlerFunc)

	log.Debugf(tag, "%c[1;0;36m [recv] msgId: %d, from sid: %d,msg: {type:%s route:%s}, execute func: %s,  params: {%+v}%c[0m",
		0x1B, msg.Id, sessionId, msgTypeBefore, routeKey, fn, reqMsg, 0x1B)

	log.Debugf(tag, "%c[1;0;36m [ret] msgId: %d, return to ri: {recvSids: %v, recvUids: %v}, msg: {type:%s route:%s}, execute func: %s, response: {%+v}%c[0m",
		0x1B, msg.Id, ri.ReceiverSids, ri.ReceiverUids, msg.Type, routeKey, fn, respMsg, 0x1B)
	return ri, isPanic
}

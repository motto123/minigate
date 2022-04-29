package server

import (
	"com.minigame.component/amqp/rabbitmq"
	"com.minigame.component/base"
	"com.minigame.component/codec/message"
	"com.minigame.component/log"
	"com.minigame.proto/chat"
	"com.minigame.proto/common"
	"com.minigame.proto/gate"
	utils "com.minigame.utils"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func sendMsgHandler(s *base.BusinessServer, sid uint64, msg *message.Message, ri *gate.ReceiveInfo) (reqMsg, respMsg proto.Message) {
	var err error
	req := new(chat.MessageReq)
	resp := &chat.MessageAck{CodeInfo: &common.ResponseCode{Code: common.Code_InternalError}}

	defer func() {
		s.DefaultErrProcess(resp.CodeInfo, err)
		msg.Type = message.Ack
		reqMsg = req
		respMsg = resp
		ri.ReceiverSids = append(ri.ReceiverSids, sid)
	}()

	if len(msg.Data) == 0 {
		resp.CodeInfo.Code = common.Code_InvalidParams
		log.Debugf(tag, "invalid params")
		return
	}

	err = proto.Unmarshal(msg.Data, req)
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf(tag, "proto.Unmarshal failed, err: %+v", err)
		return
	}

	if req.MsgId == 0 || len(req.Content) == 0 || req.SenderUid == 0 || req.ReceiverUid == 0 {
		resp.CodeInfo.Code = common.Code_InvalidParams
		log.Debugf(tag, "invalid params")
		return
	}

	resp.MsgId = req.MsgId
	resp.CodeInfo.Code = common.Code_Ok

	go func() {
		tmp := &chat.MessageNotify{
			MsgId:       req.MsgId,
			SenderUid:   req.SenderUid,
			ReceiverUid: req.ReceiverUid,
			Content:     req.Content,
		}
		marshal, err := proto.Marshal(tmp)
		if err != nil {
			err = errors.WithStack(err)
			log.Errorf(tag, "proto.Marshal failed, err: %+v", err)
			return
		}

		newMsg := message.NewMessageAndNotCompressRoute(message.Notify, message.Protobuf, msg.Id, rabbitmq.RouteReceiveMsg, marshal,
			utils.GetStructName(tmp))
		var bytes []byte
		bytes, err = newMsg.Encode()
		if err != nil {
			log.Errorf(tag, "newMsg.Encode failed, err: %+v", err)
			return
		}
		err = s.PublishDataToWriteChanel(newMsg, &gate.ReceiveInfo{
			ReceiverUids: []int64{req.ReceiverUid},
			MsgId:        newMsg.Id,
			MsgType:      int32(newMsg.Type),
			Body:         bytes,
		})
		if err != nil {
			log.Errorf(tag, "s.PublishDataToWriteChanel failed, err: %+v", err)
			return
		}
	}()
	return
}

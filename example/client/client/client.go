package client

import (
	"com.minigame.component/amqp/rabbitmq"
	"com.minigame.component/codec"
	"com.minigame.component/codec/message"
	"com.minigame.component/codec/packet"
	"com.minigame.component/codec/router"
	"com.minigame.component/log"
	"com.minigame.proto/auth"
	"com.minigame.proto/chat"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"net"
	"strconv"
	"time"
)

var (
	tag = "client"
	//readLine  = 10 * time.Minute
	//writeLine = 60 * time.Second

	readLine  = 100 * time.Minute
	writeLine = 100 * time.Minute

	idGeneral *snowflake.Node

	//routeLogin    = "login"
	//routeRegister = "register"
)

type handshakeReq struct {
	Sys struct {
		Version    string `json:"version"`
		ClientType string `json:"client_type"`
	} `json:"sys"`
	User struct{} `json:"user"`
}

type handshakeResp struct {
	Code int `json:"code"`
	Sys  struct {
		Dict map[string]uint16 `json:"dict"`
	} `json:"sys"`
	User struct{} `json:"user"`
}

type client struct {
	conn     net.Conn
	decoder  *codec.Decoder
	router   *router.Router
	uid      int64
	nickname string
	account  string
	password string
	//Ip       string
	//Port     string
}

func NewClient(Ip string, Port string) (ci client, err error) {
	ci = client{decoder: codec.NewDecoder(), router: router.NewRouter()}
	//打开连接:
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", Ip, Port))
	if err != nil {
		//由于目标计算机积极拒绝而无法创建连接
		log.Fatalf(tag, "Error dialing, %s", err.Error())
		//return // 终止程序
	}
	ci.conn = conn
	idGeneral, err = snowflake.NewNode(0)
	if err != nil {
		log.Fatalf(tag, "snowflake.NewNode, err: %+v", err.Error())
	}

	return
}

func (c *client) Do() {
	go c.readPool()
	go c.handshake()
	go c.heartbeat()

	for {
		op := ""
		arg1 := ""
		arg2 := ""
		arg3 := ""
		_, _ = fmt.Scanf("%s %s %s %s", &op, &arg1, &arg2, &arg3)
		//if err != nil {
		//	panic(err)
		//}
		if op == "" {
			//println("op is empty, please enter again ")
			continue
		}
		switch op {
		case "login":
			login(c, arg1, arg2)
		case "register":
			register(c, arg1, arg2, arg3)
		case "send":
			send(c, arg1, arg2)

		default:
			println("op is illegal, please enter again ")
		}
	}

}

func (c *client) handshake() {
	//握手
	r := handshakeReq{
		Sys: struct {
			Version    string `json:"version"`
			ClientType string `json:"client_type"`
		}{
			Version: "1.0.1",
			//Version:    "1.0.2",
			ClientType: "web",
		},
	}
	marshal, err := json.Marshal(r)
	if err != nil {
		log.Fatalf("", "json.Marshal failed, err: %+v\n", err)
		return
	}

	pkgBytes, err := c.decoder.Encode(packet.Handshake, marshal)
	if err != nil {
		log.Fatalf("", "decoder.Encode failed, err: %+v\n", err)
		return
	}

	err = c.conn.SetWriteDeadline(time.Now().Add(writeLine))
	if err != nil {
		log.Fatalf("", "decoder.Encode failed, err: %+v\n", err)
		return
	}
	_, err = c.conn.Write(pkgBytes)
	if err != nil {
		log.Fatalf("", "conn.Write failed, err: %+v\n", err)
		return
	}

}

func (c *client) heartbeat() {
	ticker := time.NewTicker(20 * time.Second)
	for {
		<-ticker.C
		//心跳
		pkgBytes, err := c.decoder.Encode(packet.Heartbeat, nil)
		if err != nil {
			log.Errorf("", "decoder.Encode failed, err: %+v\n", err)
			continue
		}

		err = c.conn.SetWriteDeadline(time.Now().Add(writeLine))
		if err != nil {
			log.Errorf("", "decoder.Encode failed, err: %+v\n", err)
			continue
		}
		_, err = c.conn.Write(pkgBytes)
		if err != nil {
			log.Errorf("", "conn.Write failed, err: %+v\n", err)
			continue
		}
	}
}

func (c *client) readPool() {
	for {
		buf1 := make([]byte, 128)
		n, err := c.conn.Read(buf1)
		if nil != err {
			log.Fatalf("", "connect.Read failed, err: %s", err)
		}
		//log.Infof("", "read buf1: %v", buf1[:n])
		var pkgs []*packet.Packet
		pkgs, err = c.decoder.Decode(buf1[:n])
		if nil != err {
			log.Fatalf("", "decoder.Decode failed, err: %s", err)
		}
		for _, pkg := range pkgs {
			err := c.handlePaket(pkg)
			if nil != err {
				//log.Errorf("", "handlePaket failed, err: %s", err)
			}
		}

	}
}

func (c *client) handlePaket(pkg *packet.Packet) error {
	switch pkg.Type {
	case packet.HandshakeAck:
		var rsp handshakeResp
		err := json.Unmarshal(pkg.Data, &rsp)
		err = errors.WithStack(err)
		if err != nil {
			return err
		}
		for k, v := range rsp.Sys.Dict {
			if k == rabbitmq.RouteLogin {
				continue
			}
			c.router.AddRouteKV("", k, v)
		}
		fmt.Printf("handshakeRsp: %+v\n", rsp)
	case packet.Kick:
		err := c.conn.Close()
		err = errors.WithStack(err)
		if err != nil {
			return err
		}

	case packet.Data:
		msg := message.NewMessageWithRouter(c.router)
		err := msg.Decode(pkg.Data)
		if nil != err {
			log.Errorf(tag, "msg.Decode failed, err: %s", err)
			return err
		}
		_ = c.handleMessage(msg)
	}
	return nil
}

func (c *client) handleMessage(msg *message.Message) (err error) {
	log.Infof(tag, "msg: %+v", msg)
	switch msg.Route {
	case rabbitmq.RouteLogin:
		var resp auth.LoginResp
		err := proto.Unmarshal(msg.Data, &resp)
		if err != nil {
			log.Errorf(tag, "proto.Unmarshal failed, err: %+v", err)
			err = errors.WithStack(err)
			return err
		}
		if resp.CodeInfo.Code != 0 {
			log.Errorf(tag, "login failed, err: %s", resp.CodeInfo.Code)
			return errors.New(resp.CodeInfo.Err)
		}
		c.uid = resp.User.Id
		c.nickname = resp.User.Nickname
		fmt.Printf("login sucessful! user: %+v", resp.User)
	case rabbitmq.RouteRegister:
		var resp auth.RegisterResp
		err := proto.Unmarshal(msg.Data, &resp)
		if err != nil {
			log.Errorf(tag, "proto.Unmarshal failed, err: %+v", err)
			err = errors.WithStack(err)
			return err
		}
		if resp.CodeInfo.Code != 0 {
			log.Errorf(tag, "register failed, err: %s", resp.CodeInfo.Code)
			return errors.New(resp.CodeInfo.Err)
		}
		fmt.Printf("regisger sucessful! uid: %d", resp.Id)
	case rabbitmq.RouteSendMsg:
		switch msg.Type {
		case message.Ack:
			var resp chat.MessageAck
			err := proto.Unmarshal(msg.Data, &resp)
			if err != nil {
				log.Errorf(tag, "proto.Unmarshal failed, err: %+v", err)
				err = errors.WithStack(err)
				return err
			}
			if resp.CodeInfo.Code != 0 {
				log.Errorf(tag, "sendMsg failed, err: %s", resp.CodeInfo.Code)
				return errors.New(resp.CodeInfo.Err)
			}
			log.Infof(tag, "[send msg ack] msgId: %d", resp.MsgId)
		case message.Notify:
			var resp chat.MessageNotify
			err := proto.Unmarshal(msg.Data, &resp)
			if err != nil {
				log.Errorf(tag, "proto.Unmarshal failed, err: %+v", err)
				err = errors.WithStack(err)
				return err
			}
			log.Infof(tag, "[snd notify] msgId: %d", resp.MsgId)
		}

	case rabbitmq.RouteReceiveMsg:
		switch msg.Type {
		case message.Ack:
			var resp chat.ReceiveAck
			err := proto.Unmarshal(msg.Data, &resp)
			if err != nil {
				log.Errorf(tag, "proto.Unmarshal failed, err: %+v", err)
				err = errors.WithStack(err)
				return err
			}
			if resp.CodeInfo.Code != 0 {
				log.Errorf(tag, "receiveMsg failed, err: %s", resp.CodeInfo.Code)
				return errors.New(resp.CodeInfo.Err)
			}
			log.Infof(tag, "[receive msg ack] msgId: %d", resp.MsgId)
		case message.Notify:
			var resp chat.MessageNotify
			err := proto.Unmarshal(msg.Data, &resp)
			if err != nil {
				log.Errorf(tag, "proto.Unmarshal failed, err: %+v", err)
				err = errors.WithStack(err)
				return err
			}
			log.Infof(tag, "[receive msg notify] msgId: %d, content:%s from %d",
				resp.MsgId, resp.Content, resp.SenderUid)

			go func() {
				receiveMsgReq(c, resp.MsgId, resp.SenderUid, resp.ReceiverUid)
			}()
		}

	}
	return err
}

func (c *client) tcpConnWrite(type1 packet.Type, msg *message.Message) (err error) {
	msgBytes, err := msg.Encode()
	if err != nil {
		log.Fatalf("", "msg.Encode failed, err: %+v\n", err)
		return
	}
	pkgBytes, err := c.decoder.Encode(type1, msgBytes)
	if err != nil {
		log.Fatalf("", "decoder.Encode failed, err: %+v\n", err)
		return
	}

	err = c.conn.SetWriteDeadline(time.Now().Add(writeLine))
	if err != nil {
		log.Fatalf("", "decoder.Encode failed, err: %+v\n", err)
		return
	}
	_, err = c.conn.Write(pkgBytes)
	if err != nil {
		log.Fatalf("", "conn.Write failed, err: %+v\n", err)
		return
	}
	return
}

func register(c *client, nickname, account, password string) {
	req := auth.RegisterReq{
		MsgId:    uint64(idGeneral.Generate().Int64()),
		Nickname: nickname,
		Account:  account,
		Password: password,
	}
	marshal, err := proto.Marshal(&req)
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf(tag, "proto.Marshal failed, err: %+v", marshal)
		return

	}
	err = c.tcpConnWrite(packet.Data, message.NewMessage(message.Request, message.Protobuf, uint64(idGeneral.Generate().Int64()),
		rabbitmq.RouteRegister, marshal, "RegisterReq", c.router))
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf(tag, "proto.Marshal failed, err: %+v", marshal)
		return
	}
}

func login(c *client, account, password string) {
	req := auth.LoginReq{
		Account:  account,
		Password: password,
	}
	marshal, err := proto.Marshal(&req)
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf(tag, "proto.Marshal failed, err: %+v", marshal)
		return

	}
	err = c.tcpConnWrite(packet.Data, message.NewMessage(message.Request, message.Protobuf, uint64(idGeneral.Generate().Int64()),
		rabbitmq.RouteLogin, marshal, "", c.router))
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf(tag, "proto.Marshal failed, err: %+v", marshal)
		return
	}
}

func send(c *client, receiver, content string) {
	receiverId, err := strconv.Atoi(receiver)
	if err != nil {
		log.Errorf(tag, "strconv.Atoi failed, err: %+v", err)
		return
	}

	req := chat.MessageReq{
		MsgId:       idGeneral.Generate().Int64(),
		Content:     content,
		SenderUid:   c.uid,
		ReceiverUid: int64(receiverId),
	}
	marshal, err := proto.Marshal(&req)
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf(tag, "proto.Marshal failed, err: %+v", marshal)
		return

	}
	err = c.tcpConnWrite(packet.Data, message.NewMessage(message.Request, message.Protobuf, uint64(idGeneral.Generate().Int64()),
		rabbitmq.RouteSendMsg, marshal, "", c.router))
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf(tag, "proto.Marshal failed, err: %+v", marshal)
		return
	}
}

func receiveMsgReq(c *client, msgId, senderUid, receiverUid int64) {
	req := chat.ReceiveReq{
		MsgId:       msgId,
		SenderUid:   senderUid,
		ReceiverUid: receiverUid,
	}
	marshal, err := proto.Marshal(&req)
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf(tag, "proto.Marshal failed, err: %+v", marshal)
		return

	}
	err = c.tcpConnWrite(packet.Data, message.NewMessage(message.Request, message.Protobuf, uint64(idGeneral.Generate().Int64()),
		rabbitmq.RouteReceiveMsg, marshal, "", c.router))
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf(tag, "proto.Marshal failed, err: %+v", marshal)
		return
	}
}

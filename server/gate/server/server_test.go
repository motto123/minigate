package server

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"com.minigame.component/codec"
	"com.minigame.component/codec/message"
	"com.minigame.component/codec/packet"
	"com.minigame.component/codec/router"
	"com.minigame.component/log"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

var (
	readLine  = 10 * time.Minute
	writeLine = 60 * time.Second
	decoder   = codec.NewDecoder()
	newRouter = router.NewRouter()
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

func TestConn(t *testing.T) {
	log.InitLogger("server_test", "./")
	createTcpClient()
	//createWsClient()

}

func createWsClient() {
	newRouter := router.NewRouter()
	newRouter.AddRouteKV("test", 9)
	newRouter.AddRouteKV("login", 10)
	decoder := codec.NewDecoder()
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:5555/web", nil)

	if err != nil {
		//由于目标计算机积极拒绝而无法创建连接
		log.Fatalf("test", "Error dialing: %s", err.Error())
		return // 终止程序
	}
	defer conn.Close()

	conn.SetCloseHandler(func(code int, text string) error {
		log.Debugf("test", "code: %d, text: %s", code, text)
		return nil
	})
	//err = conn.Close()
	//return

	//msg := message.Message{
	//	//Type:     message.Ack,
	//	Type:     message.Ack,
	//	DataType: message.Protobuf,
	//	Id:       9328453,
	//	Route:    "test",
	//	Data:     []byte("0123456789abcd"),
	//	//Data:        []byte("3"),
	//	DataObjName: "A",
	//}
	//msg.SetRouter(newRouter)
	msg := message.NewMessage(
		message.Ack,
		message.Protobuf,
		9328453,
		"test",
		[]byte("0123456789abcd"),
		"A",
		newRouter,
	)
	msgByts, err := msg.Encode()
	if err != nil {
		log.Fatalf("", "msg.Encode failed, err: %S", err)
	}
	log.Infof("", "write msgByts: %v", msgByts)
	//log.Infof("", "write msgByts: %s", msgByts)

	buf, err := decoder.Encode(packet.Data, msgByts)
	if err != nil {

		log.Fatalf("", "decoder.Encode failed, err: %S", err)
	}
	log.Infof("", "write buf: %v", buf)
	//log.Infof("", "write buf: %s", buf)
	err = conn.WriteMessage(websocket.BinaryMessage, buf)
	if err != nil {
		log.Fatalf("", "conn.WriteMessage failed, err: %S", err)
	}

	for {
	}
}

func createTcpClient() {
	//打开连接:
	conn, err := net.Dial("tcp", "localhost:5554")
	if err != nil {
		//由于目标计算机积极拒绝而无法创建连接
		log.Fatalf("", "Error dialing, %s", err.Error())
		//return // 终止程序
	}
	defer conn.Close()

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

	pkgBytes, err := decoder.Encode(packet.Handshake, marshal)
	if err != nil {
		log.Fatalf("", "decoder.Encode failed, err: %+v\n", err)
		return
	}

	err = conn.SetWriteDeadline(time.Now().Add(writeLine))
	if err != nil {
		log.Fatalf("", "decoder.Encode failed, err: %+v\n", err)
		return
	}
	_, err = conn.Write(pkgBytes)
	if err != nil {
		log.Fatalf("", "conn.Write failed, err: %+v\n", err)
		return
	}

	//心跳
	pkgBytes, err = decoder.Encode(packet.Heartbeat, nil)
	if err != nil {
		log.Fatalf("", "decoder.Encode failed, err: %+v\n", err)
		return
	}

	err = conn.SetWriteDeadline(time.Now().Add(writeLine))
	if err != nil {
		log.Fatalf("", "decoder.Encode failed, err: %+v\n", err)
		return
	}
	_, err = conn.Write(pkgBytes)
	if err != nil {
		log.Fatalf("", "conn.Write failed, err: %+v\n", err)
		return
	}

	for {
		buf1 := make([]byte, 128)
		n, err := conn.Read(buf1)
		if nil != err {
			log.Fatalf("", "connect.Read failed, err: %s", err)
		}
		log.Infof("", "read buf1: %v", buf1[:n])
		var pkgs []*packet.Packet
		pkgs, err = decoder.Decode(buf1[:n])
		if nil != err {
			log.Fatalf("", "decoder.Decode failed, err: %s", err)
		}
		for _, pkg := range pkgs {
			err := handlePaket(conn, pkg)
			if nil != err {
				log.Errorf("", "handlePaket failed, err: %s", err)
			}
		}

	}
}

func tcpConnWrite(conn net.Conn, type1 packet.Type, msg *message.Message) (err error) {
	msgBytes, err := msg.Encode()
	if err != nil {
		log.Fatalf("", "msg.Encode failed, err: %+v\n", err)
		return
	}
	pkgBytes, err := decoder.Encode(type1, msgBytes)
	if err != nil {
		log.Fatalf("", "decoder.Encode failed, err: %+v\n", err)
		return
	}

	err = conn.SetWriteDeadline(time.Now().Add(writeLine))
	if err != nil {
		log.Fatalf("", "decoder.Encode failed, err: %+v\n", err)
		return
	}
	_, err = conn.Write(pkgBytes)
	if err != nil {
		log.Fatalf("", "conn.Write failed, err: %+v\n", err)
		return
	}
	return
}

func handlePaket(conn net.Conn, pkg *packet.Packet) error {
	switch pkg.Type {
	case packet.HandshakeAck:
		var rsp handshakeResp
		err := json.Unmarshal(pkg.Data, &rsp)
		err = errors.WithStack(err)
		if err != nil {
			return err
		}
		for k, v := range rsp.Sys.Dict {
			newRouter.AddRouteKV(k, v)
		}

		msg := message.NewMessage(
			message.Request,
			message.Protobuf,
			9328453,
			"AAAAAAAA",
			[]byte("3"),
			"",
			newRouter,
		)
		err = tcpConnWrite(conn, packet.Data, msg)
		if err != nil {
			return err
		}

		msg.Route = "test"
		err = tcpConnWrite(conn, packet.Data, msg)
		if err != nil {
			return err
		}

	case packet.Kick:
		err := conn.Close()
		err = errors.WithStack(err)
		if err != nil {
			return err
		}

	case packet.Data:
		msg := message.NewMessageWithRouter(newRouter)
		err := msg.Decode(pkg.Data)
		if nil != err {
			log.Fatalf("", "msg.Decode failed, err: %s", err)
		}
		log.Infof("", "msg: %+v", msg)
	}
	return nil
}

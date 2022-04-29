package session

import (
	"fmt"
	"net"
	"time"

	"com.minigame.component/codec"
	"com.minigame.component/codec/packet"
	"com.minigame.component/codec/router"
	"com.minigame.component/log"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

const (
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 64 * 1024
)

type WsSession struct {
	Id                   uint64
	Conn                 *websocket.Conn
	NetProtocol          string
	decoder              *codec.Decoder
	unregister           chan uint64
	Send                 chan *PostMsg
	router               *router.Router
	IsCheckHeartbeat     bool
	IsRecordHeartbeatLog bool

	Uid      int64
	Nickname string
}

func (s *WsSession) Init() {
	tag = fmt.Sprintf("%s_%s", s.NetProtocol, tag)

	s.Conn.SetReadLimit(maxMessageSize)
	s.Conn.SetPingHandler(func(appData string) error {
		log.Infof(tag, "Conn.SetPingHandler, appData: %s", appData)
		return nil
	})
	s.Conn.SetPongHandler(func(string) error {
		log.Infof(tag, "Conn.SetPongHandler")
		s.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	s.Conn.SetCloseHandler(func(code int, text string) error {
		log.Infof(tag, "SetCloseHandler, code: %d, text: %s\n", code, text)
		//s.unregister <- s.Id
		return nil
	})

}

func (s *WsSession) GetId() uint64 {
	return s.Id
}

func (s *WsSession) GetDecoder() *codec.Decoder {
	return s.decoder
}

func (s *WsSession) GetSend() chan *PostMsg {
	return s.Send
}

func (s *WsSession) GetRouter() *router.Router {
	return s.router
}

func (s *WsSession) Write(data []byte) (int, error) {
	err := s.Conn.WriteMessage(websocket.BinaryMessage, data)
	err = errors.WithStack(err)
	return len(data), err
}

func (s *WsSession) Close() error {
	return errors.WithStack(s.Conn.Close())
}

func (s *WsSession) LocalAddr() net.Addr {
	return s.Conn.LocalAddr()
}

func (s *WsSession) RemoteAddr() net.Addr {
	return s.Conn.RemoteAddr()
}

func (s *WsSession) SetReadDeadline(t time.Time) error {
	return errors.WithStack(s.Conn.SetReadDeadline(t))
}

func (s *WsSession) SetWriteDeadline(t time.Time) error {
	return errors.WithStack(s.Conn.SetWriteDeadline(t))
}

func (s *WsSession) GetRecordHeartbeatLog() bool {
	return s.IsRecordHeartbeatLog
}

func (s *WsSession) SetCustomData(uid int64, nickname string) {
	s.Uid = uid
	s.Nickname = nickname
}

func (s *WsSession) GetUid() int64 {
	return s.Uid
}

func NewWsSession(id uint64, conn *websocket.Conn, unregister chan uint64, send chan *PostMsg,
	isCheckHeartbeat, isRecordHeartbeatLog bool) *WsSession {
	return &WsSession{
		Id:                   id,
		Conn:                 conn,
		NetProtocol:          WEBSOCKET,
		decoder:              codec.NewDecoder(),
		unregister:           unregister,
		Send:                 send,
		IsCheckHeartbeat:     isCheckHeartbeat,
		IsRecordHeartbeatLog: isRecordHeartbeatLog,
	}
}

func (s *WsSession) ReadPump() {
	go func() {
		defer func() {
			s.unregister <- s.Id
		}()
		for {
			err := s.Conn.SetReadDeadline(time.Now().Add(readLine))
			if err != nil {
				log.Errorf(tag, "Conn.SetReadDeadline failed, err: %s", err)
				return
			}

			t, message, err := s.Conn.ReadMessage()
			err = errors.WithStack(err)
			if err != nil {
				if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Debugf(tag, "websocket.IsUnexpectedCloseError, error: %+v", err)
					return
				}
				log.Errorf(tag, "Conn.ReadMessage, error: %+v", err)
			}
			log.Infof(tag, "ReadMessage,t: %d,  message: %+v", t, message)
			switch t {
			case websocket.TextMessage:
			case websocket.BinaryMessage:
				var pkgList []*packet.Packet
				pkgList, err := s.decoder.Decode(message)
				if err != nil {
					log.Errorf(tag, "decoder.Decode, error: %+v", err)
					continue
				}
				for _, pkg := range pkgList {
					err = handlePacket(s, pkg)
					if err != nil {
						log.Errorf(tag, "handlePacket, error: %+v", err)
						continue
					}
				}
			}
		}
	}()
}

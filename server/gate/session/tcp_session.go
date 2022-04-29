package session

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"com.minigame.component/codec"
	"com.minigame.component/codec/packet"
	"com.minigame.component/codec/router"
	"com.minigame.component/log"
	"github.com/pkg/errors"
)

type TcpSession struct {
	Id                   uint64
	Conn                 net.Conn
	NetProtocol          string
	decoder              *codec.Decoder
	unregister           chan uint64
	Send                 chan *PostMsg
	router               *router.Router
	lastPackageTime      int64
	IsCheckHeartbeat     bool
	IsRecordHeartbeatLog bool

	Uid      int64
	Nickname string
}

func (s *TcpSession) Init() {
	tag = fmt.Sprintf("%s_%s", s.NetProtocol, tag)
}

func (s *TcpSession) GetId() uint64 {
	return s.Id
}

func (s *TcpSession) GetDecoder() *codec.Decoder {
	return s.decoder
}

func (s *TcpSession) GetSend() chan *PostMsg {
	return s.Send
}

func (s *TcpSession) GetRouter() *router.Router {
	return s.router
}

func (s *TcpSession) Write(data []byte) (int, error) {
	// 更新上一个包的时间
	s.lastPackageTime = time.Now().Unix()
	n, err := s.Conn.Write(data)
	err = errors.WithStack(err)
	return n, err
}

func (s *TcpSession) Close() error {
	return errors.WithStack(s.Conn.Close())
}

func (s *TcpSession) LocalAddr() net.Addr {
	return s.Conn.LocalAddr()
}

func (s *TcpSession) RemoteAddr() net.Addr {
	return s.Conn.RemoteAddr()
}

func (s *TcpSession) SetDeadline(t time.Time) error {
	return errors.WithStack(s.SetDeadline(t))
}

func (s *TcpSession) SetReadDeadline(t time.Time) error {
	return errors.WithStack(s.Conn.SetReadDeadline(t))
}

func (s *TcpSession) SetWriteDeadline(t time.Time) error {
	return errors.WithStack(s.Conn.SetWriteDeadline(t))
}

func (s *TcpSession) GetRecordHeartbeatLog() bool {
	return s.IsRecordHeartbeatLog
}

func (s *TcpSession) SetCustomData(uid int64, nickname string) {
	s.Uid = uid
	s.Nickname = nickname
}

func (s *TcpSession) GetUid() int64 {
	return s.Uid
}

func NewTcpSession(id uint64, conn net.Conn, unregister chan uint64, send chan *PostMsg, router *router.Router,
	isCheckHeartbeat, isRecordHeartbeatLog bool) *TcpSession {
	return &TcpSession{
		Id:                   id,
		Conn:                 conn,
		NetProtocol:          TCP,
		decoder:              codec.NewDecoder(),
		unregister:           unregister,
		Send:                 send,
		router:               router,
		IsCheckHeartbeat:     isCheckHeartbeat,
		IsRecordHeartbeatLog: isRecordHeartbeatLog,
	}
}

func (s *TcpSession) ReadPump() {
	go func() {
		defer func() {
			s.unregister <- s.Id
		}()

		var err error
		buf := make([]byte, 1024)
		for {
			if s.IsCheckHeartbeat {
				err = s.SetReadDeadline(time.Now().Add(readLine))
				if err != nil {
					log.Errorf(tag, "session.SetReadDeadline, error: %+v", err)
					return
				}
			}
			var n int
			n, err = s.Conn.Read(buf)
			if err != nil {
				err = errors.WithStack(err)
				if errors.Is(err, io.EOF) {
					return
				}
				log.Errorf(tag, "Conn.Read, error: %+v", err)
				continue
			}
			// 更新上一个包的时间
			s.lastPackageTime = time.Now().Unix()

			//if n == 0 {
			//	continue
			//}
			//fmt.Printf("read buf: %v\n", buf[:n])
			//fmt.Printf("read buf: %s\n", buf[:n])

			var pkgList []*packet.Packet
			pkgList, err := s.decoder.Decode(buf[:n])
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
	}()
}

func handlePacket(s Session, pkg *packet.Packet) (err error) {
	switch pkg.Type {
	case packet.Handshake:
		//握手逻辑
		if pkg.Length == 0 {
			return
		}
		err = handshake(s, pkg.Data)
		if err != nil {
			return
		}

	case packet.Heartbeat:
		if s.GetRecordHeartbeatLog() {
			log.Debugf(tag, "receive session %d heartbeat", s.GetId())
		}
	case packet.Data:
		if len(pkg.Data) == 0 {
			log.Debugf(tag, "receive session %d is empty msg data", s.GetId())
			return
		}

		s.GetSend() <- &PostMsg{
			SessionId:    s.GetId(),
			MsgBytesList: [][]byte{pkg.Data},
		}
	}

	return
}

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

func handshake(s Session, data []byte) (err error) {
	var r handshakeReq
	err = json.Unmarshal(data, &r)
	err = errors.WithStack(err)
	if err != nil {
		return
	}
	if r.Sys.Version != "1.0.1" {
		err = errors.New("version match failed")
		return
	}

	ret := handshakeResp{
		Code: 200,
		Sys: struct {
			Dict map[string]uint16 `json:"dict"`
		}{
			Dict: s.GetRouter().GetRoutes(),
		},
		User: struct{}{},
	}
	retBytes, err := json.Marshal(ret)
	err = errors.WithStack(err)
	if err != nil {
		return
	}
	buf, err := s.GetDecoder().Encode(packet.HandshakeAck, retBytes)
	if err != nil {
		return
	}

	err = s.SetWriteDeadline(time.Now().Add(WriteLine))
	_, err = s.Write(buf)
	if err != nil {
		return
	}
	return
}

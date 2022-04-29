package session

import (
	"com.minigame.component/codec"
	"com.minigame.component/codec/router"

	//"com.minigame.proto/gate"
	"net"
	"time"
)

const (
	TCP       = "tcp"
	WEBSOCKET = "ws"

	readLine  = 10 * time.Minute
	WriteLine = 60 * time.Second

	//readLine  = 5 * time.Second
	//WriteLine = 5 * time.Second
)

var tag = "session"

type PostMsg struct {
	SessionId    uint64
	MsgBytesList [][]byte
}

type Session interface {
	Init()
	GetId() uint64
	GetDecoder() *codec.Decoder
	GetSend() chan *PostMsg
	GetRouter() *router.Router
	Write(data []byte) (int, error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	GetRecordHeartbeatLog() bool
	SetCustomData(uid int64, nickname string)
	GetUid() int64
}

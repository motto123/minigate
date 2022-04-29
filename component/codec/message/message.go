package message

import (
	"bytes"

	"com.minigame.component/codec/router"
	"github.com/pkg/errors"
)

// Message types
// s: server
// c: client
// who can use some message type
const (
	Request  Type = 0x00 // c
	Push          = 0x01 // c
	Response      = 0x02 // s
	Notify        = 0x03 // s
	Ack           = 0x04 // s
)

var types = map[Type]string{
	Request:  "Request",
	Notify:   "Notify",
	Response: "Response",
	Push:     "Push",
	Ack:      "Ack",
}

const (
	Text     DataType = 0x00
	Json     DataType = 0x01
	Protobuf DataType = 0x02
)

var dataTypes = map[DataType]string{
	Text:     "text",
	Json:     "json",
	Protobuf: "protobuf",
}

const (
	flagLen    = 1
	msgLen     = 1
	routeLen   = 1
	objNameLen = 1
)

const (
	msgRouteCompressMask = 0x01
	msgDataTypeMask      = 0x03
	msgTypeMask          = 0x07
)

var (
	ErrWrongMessageType = errors.New("wrong message type")
	ErrWrongDataType    = errors.New("wrong data type")
	ErrInvalidMessage   = errors.New("invalid message")
	//ErrRouteInfoNotFound = errors.New("route info not found in dictionary")
	//ErrWrongMessage      = errors.New("wrong message")
)

type Message struct {
	Type        Type // message type
	compressed  bool // whether compress route
	DataType    DataType
	Id          uint64 // unique id, zero while notify mode, 8bytes max 18446744073709551615
	Route       string // route for locating service
	routeCode   uint16
	Data        []byte // payload
	DataObjName string
	router      *router.Router
}

// NewMessage 创建普通的message
func NewMessage(type1 Type, dataType DataType, id uint64, route string, data []byte, dataObjName string,
	router *router.Router) *Message {
	return &Message{
		Type:        type1,
		compressed:  true,
		DataType:    dataType,
		Id:          id,
		Route:       route,
		Data:        data,
		DataObjName: dataObjName,
		router:      router,
	}
}

func NewMessageWithRouter(router *router.Router) *Message {
	return NewMessage(0, 0, 0, "", nil, "", router)
}

// NewMessageAndNotCompressRoute 用于不处理压缩路由的message
func NewMessageAndNotCompressRoute(type1 Type, dataType DataType, id uint64, route string, data []byte, dataObjName string) *Message {
	return &Message{
		Type:        type1,
		compressed:  false,
		DataType:    dataType,
		Id:          id,
		Route:       route,
		Data:        data,
		DataObjName: dataObjName,
		router:      router.NewRouter(),
	}
}

// NewMessageAndUnchangedRoute 用于保持路由数据不变的message
func NewMessageAndUnchangedRoute() *Message {
	return &Message{}
}

// setAutoCompressed 是否自动压缩路由,测试用
func (m *Message) setAutoCompressed(b bool) {
	m.compressed = b
}

// Decode binary to message format.
// 压缩路由有3种状态,1压缩 2不压缩 3保持不变(被压缩,但是没有路由字典表)
func (m *Message) Decode(data []byte) error {
	if len(data) < flagLen {
		return ErrInvalidMessage
	}

	buf := bytes.NewBuffer(data)
	flagByt := buf.Next(flagLen)
	f0 := flagByt[0]
	// 取出末尾3bits的数据
	m.Type = Type((f0 >> 3) & msgTypeMask)
	if invalidType(m.Type) {
		return ErrWrongMessageType
	}
	m.DataType = DataType(f0 & msgDataTypeMask)

	// 取出末尾1bit的数据,看最低为是否为1
	m.compressed = (f0 >> 2 & msgRouteCompressMask) == 1

	if buf.Len() > 0 {
		msgIdLen := int(buf.Next(msgLen)[0])
		if msgIdLen > 0 && buf.Len() >= msgIdLen {
			next := buf.Next(msgIdLen)
			m.Id = bytesToUint64(next)
		}
	}
	if buf.Len() > 0 {
		routeLen := int(buf.Next(routeLen)[0])
		if routeLen > 0 && buf.Len() >= routeLen {
			next := buf.Next(routeLen)
			if m.compressed && m.router == nil { //保持不变,存储压缩后的code
				routeCode := uint16(bytesToUint64(next))
				m.routeCode = routeCode
			} else if m.compressed { //路由压缩了
				routeCode := uint16(bytesToUint64(next))
				route, ok := m.router.GetRouteName(routeCode)
				if !ok { //不在压缩路由字典里,不压缩
					m.Route = string(next)
					m.compressed = false
				} else {
					m.Route = route
				}
			} else { //没有压缩的路由
				m.Route = string(next)
			}
		}
	}

	if buf.Len() > 0 {
		objNameLen := int(buf.Next(objNameLen)[0])
		if objNameLen > 0 && buf.Len() >= objNameLen {
			m.DataObjName = string(buf.Next(objNameLen))
		}
	}

	if buf.Len() > 0 {
		m.Data = buf.Bytes()
	}
	return nil
}

// Encode message to binary format.
// 压缩路由有3种状态,1压缩 2不压缩 3保持不变(被压缩,但是没有路由字典表)
func (m *Message) Encode() ([]byte, error) {
	if invalidType(m.Type) {
		return nil, ErrWrongMessageType
	}
	if invalidDataType(m.DataType) {
		return nil, ErrWrongDataType
	}
	buf := make([]byte, 0)
	// set message type
	flagByt := byte(m.Type)
	// set route tag
	flagByt = flagByt << 1
	if m.compressed {
		// 把末尾1bit,设置为1
		flagByt |= msgRouteCompressMask
	}
	// set data type
	b := flagByt << 2
	flagByt = b | byte(m.DataType)
	// add flag into the buf
	buf = append(buf, flagByt)

	tempBuf := uint64ToBytes(m.Id)
	if m.Id != 0 {
		// add message id len into the buf
		buf = append(buf, byte(len(tempBuf)))
		// add message id  into the buf
		buf = append(buf, tempBuf...)
	} else {
		buf = append(buf, byte(0))
	}

	// compressed route
	if m.compressed && m.router == nil {
		toBytes := uint64ToBytes(uint64(m.routeCode))
		buf = append(buf, byte(len(toBytes)))
		//add route into the buf
		buf = append(buf, toBytes...)
	} else if m.compressed {
		if len(m.Route) != 0 {
			//add route len into the buf
			u, ok := m.router.GetRouteCode(m.Route)
			if !ok { //不在压缩路由字典里
				buf = append(buf, byte(len(m.Route)))
				buf = append(buf, []byte(m.Route)...)
				//Oxfb equal to 1111 1011,把第6位重置为0
				m.compressed = false
				buf[0] &= 0xfb
			} else {
				toBytes := uint64ToBytes(uint64(u))
				buf = append(buf, byte(len(toBytes)))
				//add route into the buf
				buf = append(buf, toBytes...)
			}
		} else {
			buf[0] &= 0xfb
			buf = append(buf, byte(0))
		}
	} else {
		// don't compressed route
		if len(m.Route) != 0 {
			// add route len into the buf
			buf = append(buf, byte(len(m.Route)))
			// add route into the buf
			buf = append(buf, []byte(m.Route)...)
		} else {
			buf = append(buf, byte(0))
		}
	}

	buf = append(buf, byte(len(m.DataObjName)))
	if len(m.DataObjName) != 0 {
		// add obj name into the buf
		buf = append(buf, []byte(m.DataObjName)...)
	}

	buf = append(buf, m.Data...)
	return buf, nil
}

func (m *Message) ErasureDataInfoWithoutOther() {
	//m.Id
	//m.Route
	//m.routeCode
	//m.compressed
	//m.router
	//m.Type = 0

	m.DataType = 0
	m.Data = nil
	m.DataObjName = ""
}

func invalidType(t Type) bool {
	return t < Request || t > Ack
}

func invalidDataType(t DataType) bool {
	return t < Text || t > Protobuf
}

// 大端模式
// if receive  0001 0000 0010 return 258
func bytesToUint64(b []byte) uint64 {
	result := uint64(0)
	for _, v := range b {
		u := result << 8
		result = u + uint64(v)
	}
	return result
}

// 大端模式
// as 258 return  0001 0000 0010
func uint64ToBytes(n uint64) []byte {
	buf := make([]byte, 0)
	if n == 0 {
		buf = append(buf, uint8(0))
		return buf
	}
	for i := 0; i < 8; i++ {
		buf = append(buf, uint8(n))
		n >>= 8
	}
	tempBuf := make([]byte, 0)
	tag := 0
	for i := 0; i < len(buf); i++ {
		b := buf[len(buf)-i-1]
		if b == 0x0 && tag == 0 {
			continue
		} else {
			tag++
		}
		tempBuf = append(tempBuf, b)
	}

	return tempBuf
}

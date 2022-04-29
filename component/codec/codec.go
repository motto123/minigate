package codec

import (
	"bytes"
	"com.minigame.component/codec/packet"
	"github.com/pkg/errors"
)

const (
	headLen       = 5
	MaxPacketSize = 64 * 1024
)

var ErrPacketSizeExceed = errors.New("codec: packet size exceed")
var ErrDataHeadLenShort = errors.New("codec: data head len is short")

// Decoder 数据协议格式 Header(Type-Len)-Body(Data),Head必须是4个字节,Body最大64kb
type Decoder struct {
	buf *bytes.Buffer
	//是否是组合包
	hasCombination bool
	combinationLen int
}

func NewDecoder() *Decoder {
	return &Decoder{buf: bytes.NewBuffer(nil)}
}

func (d *Decoder) forward() (typ byte, len int, err error) {
	if d.buf.Len() < headLen {
		return 0, 0, ErrDataHeadLenShort
	}
	header := d.buf.Next(headLen)
	typ = header[0]
	if typ < packet.Handshake || typ > packet.Kick {
		err = packet.ErrWrongPacketType
		return
	}
	len = bytesToInt(header[1 : headLen-1])
	if len > MaxPacketSize {
		err = ErrPacketSizeExceed
		return
	}

	if len > d.buf.Len() {
		//要组包,说明从conn.read里还有不完整的数据包,等待读取出一个完整的数据包在处理
		d.hasCombination = true
		d.combinationLen = len
		tempbyts := make([]byte, d.buf.Len())
		copy(tempbyts, d.buf.Bytes())
		d.buf.Reset()
		d.buf.Write(header)
		d.buf.Write(tempbyts)
	}
	return
}

func (d *Decoder) Decode(data []byte) (packets []*packet.Packet, err error) {
	if d.hasCombination { //组合数据包,小概率
		_, err = d.buf.Write(data)
		err = errors.WithStack(err)
		if err != nil {
			return
		}
		//因为header被重新写入了,所以要buf.len -head len
		//if d.buf.Len()-headLen < d.combinationLen {
		if d.buf.Len() < d.combinationLen {
			return nil, nil
		}
		d.hasCombination = false
	} else {
		d.buf.Write(data)
	}

continueDecode:
	typ, len, err := d.forward()
	if err != nil {
		return nil, err
	}

	if d.hasCombination {
		return
	}

	p := &packet.Packet{
		Type:   packet.Type(typ),
		Length: len,
		Data:   d.buf.Next(len),
	}

	packets = append(packets, p)
	d.combinationLen = 0

	if d.buf.Len() >= headLen {
		//发生沾包,继续decode
		goto continueDecode
	}
	return
}

func (d *Decoder) Encode(typ packet.Type, data []byte) (buf []byte, err error) {
	if typ < packet.Handshake || typ > packet.Kick {
		return nil, packet.ErrWrongPacketType
	}
	p := &packet.Packet{
		Type:   typ,
		Length: len(data),
		Data:   data,
	}
	buf = make([]byte, p.Length+headLen)
	buf[0] = byte(typ)
	copy(buf[1:headLen-1], intToBytes(p.Length))
	//TODO:拆包逻辑,没实现
	buf[headLen-1] = byte(0)
	copy(buf[headLen:], data)
	return
}

// 大端模式 bytes transfer int
func bytesToInt(b []byte) int {
	result := 0
	for _, v := range b {
		result = result<<8 + int(v)
	}
	return result
}

// 大端模式 int transfer bytes
func intToBytes(n int) []byte {
	buf := make([]byte, 3)
	buf[0] = byte((n >> 16) & 0xFF)
	buf[1] = byte((n >> 8) & 0xFF)
	buf[2] = byte(n & 0xFF)
	return buf
}

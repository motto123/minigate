package codec

import (
	"bytes"
	"com.minigame.component/codec/packet"
	"testing"
)

func TestDecodeSinglePacket(t *testing.T) {
	decoder := NewDecoder()
	data := make([]byte, 258)
	encode, err := decoder.Encode(packet.Data, data)
	if err != nil {
		t.Fatal(err)
	}
	decode, err := decoder.Decode(encode)
	if err != nil {
		t.Fatal(err)
	}
	if len(decode) != 1 {
		t.Fatal("packets len dont equal to 1")
	}
	if !(decode[0].Type == packet.Data && string(decode[0].Data) == string(data)) {
		t.Fatal("decode failed")
	}
}

func TestDecodeMultiPacket(t *testing.T) {
	decoder := NewDecoder()
	data := []byte("hehe")
	encode, err := decoder.Encode(packet.Data, data)
	if err != nil {
		t.Fatal(err)
	}
	//t.Logf("encode: %v\n", encode)
	buf := bytes.NewBuffer(encode)
	buf.Write(encode)
	buf.Write(encode)
	//t.Logf("buf: %v\n", buf.Bytes())

	decode, err := decoder.Decode(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(decode) != 3 {
		t.Fatal("packets len dont equal to 3")
	}
	for _, p := range decode {
		if !(p.Type == packet.Data && string(p.Data) == string(data)) {
			t.Fatal("decode failed")
		}
	}
}

func TestDecodeCombinationPacket(t *testing.T) {
	var fibTests = []struct {
		in       string // input
		expected string // expected result
	}{
		//[head bod,y] [head body] 单个包长9
		{"abab", "abab"},
		//[head body], [head body] 单个包长10
		{"ababa", "ababa"},
		//[head body] [h,ead body] 单个包长11
		{"123456", "123456"},
	}

	for _, tt := range fibTests {
		decoder := NewDecoder()
		data := []byte(tt.in)
		encode, err := decoder.Encode(packet.Data, data)
		if err != nil {
			t.Fatal(err)
		}
		//fmt.Printf("encode: %s\n", encode)

		i := 1
		var pkgs []*packet.Packet
		temp := encode
		temp = append(encode, encode...)
		//fmt.Printf("temp: %v, len: %d\n", temp, len(temp))
		for len(temp)-10*i == 0 || len(temp)-10*i > -10 {
			end := i * 10
			if len(temp)-10*i < 0 {
				end = end + (len(temp) - 10*i)
			}
			d := temp[(i-1)*10 : end]
			//fmt.Printf("d: %v\n", d)
			//fmt.Printf("(i-1)*10:%d, end: %d\n", (i-1)*10, end)
			arr, err := decoder.Decode(d)
			if err != nil {
				t.Fatal(err)
			}
			i++
			if arr != nil {
				pkgs = append(pkgs, arr...)

			}

		}
		if len(pkgs) == 0 {
			t.Fatal("decode failed, because pkgs len equal to 0")
		}

		for i := range pkgs {
			p := pkgs[i]
			if !(p.Type == packet.Data && string(p.Data) == tt.expected) {
				t.Errorf("decode failed, expect: %s, but actual: %s", tt.in, p.Data)
			}
		}
	}
}

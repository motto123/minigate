package message

import (
	"com.minigame.component/codec/router"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestMessageType(t *testing.T) {
	for i := 0; i < 6; i++ {
		t := Type(i)
		fmt.Printf("type: %s, %v, %d\n", t, t, t)
	}
}

func TestUintToBytes(t *testing.T) {
	arr := []uint64{66052788, 2048, 2064, 258, 512, 1, 0, 10, 9, 8, 0}
	for _, n := range arr {
		toBytes := uint64ToBytes(n)
		toUint64 := bytesToUint64(toBytes)
		//if n != toUint64 {
		//	t.Fatalf("expect %d, actual: %d", n, toUint64)
		//}
		assert.Equal(t, n, toUint64)
	}

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100000000; i++ {
		n := uint64(rand.Uint32())
		toBytes := uint64ToBytes(n)
		toUint64 := bytesToUint64(toBytes)
		//if n != toUint64 {
		//	t.Fatalf("expect %d, actual: %d", n, toUint64)
		//}
		assert.Equal(t, n, toUint64)
	}
}

func TestMessage(t *testing.T) {
	newRouter := router.NewRouter()
	newRouter.AddRouteKV("1", "login", 1)
	newRouter.AddRouteKV("auth", "register", 2)
	m := NewMessage(Ack, Protobuf, 512, "test", []byte("abc"), "Person", newRouter)
	mByts, err := m.Encode()
	if err != nil {
		t.Fatalf("err: %+v", err)
	}
	fmt.Println(mByts)

	m1 := NewMessageWithRouter(newRouter)
	err = m1.Decode(mByts)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Printf("m: %+v\n", m)
	//fmt.Printf("m1: %+v\n", m1)
	//fmt.Printf("%v\n", m1.Data)

	assert.Equal(t, m1, m)

	//byts := []byte{6, 1, 1, 1, 2, 0}
	//msg := NewMessageWithRouter(newRouter)
	//err = msg.Decode(byts)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Printf("msg: %+v\n", msg)
}
func TestAutoCompressedRouter(t *testing.T) {
	//自动压缩路由
	arr := []string{"A", "B", "C"}

	newRouter := router.NewRouter()
	for i, v := range arr {
		//i += 10
		newRouter.AddRouteKV("1", v, uint16(i))
	}
	for _, v := range arr {
		m := NewMessage(Notify, 0, 0, v, nil, "", newRouter)
		mByts, err := m.Encode()
		if err != nil {
			t.Fatalf("err: %+v", err)
		}
		fmt.Println(mByts)
		assert.True(t, m.compressed)

		m1 := Message{
			router: newRouter,
		}
		err = m1.Decode(mByts)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(mByts)
		assert.True(t, m1.compressed)
		assert.Equal(t, m.Route, m1.Route)
	}
}

func TestMessageWithRouter(t *testing.T) {
	arr := []string{"A", "B", "C"}
	newRouter := router.NewRouter()
	for i, v := range arr {
		//i += 10
		newRouter.AddRouteKV("1", v, uint16(i))
	}
	//自动压缩路由失败，转为不压缩
	m := NewMessage(Notify, 0, 0, "AA", nil, "", newRouter)
	mByts, err := m.Encode()
	if err != nil {
		t.Fatalf("err: %+v", err)
	}
	//fmt.Println(mByts)
	assert.False(t, m.compressed)

	m1 := NewMessageWithRouter(newRouter)
	err = m1.Decode(mByts)
	if err != nil {
		t.Fatal(err)
	}
	myByts1, err := m1.Encode()
	if err != nil {
		t.Fatalf("err: %+v", err)
	}
	//fmt.Println(mByts)
	assert.False(t, m1.compressed)
	assert.Equal(t, m.Route, m1.Route)
	assert.Equal(t, mByts, myByts1)

	//不压缩路由
	m = NewMessage(Notify, 0, 0, "BB", nil, "", newRouter)
	m.setAutoCompressed(false)
	mByts, err = m.Encode()
	if err != nil {
		t.Fatalf("err: %+v", err)
	}
	//fmt.Println(mByts)
	assert.False(t, m.compressed)

	m1 = NewMessageWithRouter(newRouter)
	err = m1.Decode(mByts)
	if err != nil {
		t.Fatal(err)
	}
	myByts1, err = m1.Encode()
	if err != nil {
		t.Fatalf("err: %+v", err)
	}
	//fmt.Println(mByts)
	assert.Equal(t, mByts, myByts1)
	assert.False(t, m1.compressed)
	assert.Equal(t, m1, m)

	//
	m = NewMessage(Notify, 1, 1, "BB", nil, "", nil)
	mByts, err = m.Encode()
	if err != nil {
		t.Fatalf("err: %+v", err)
	}
	fmt.Println(mByts)
	assert.False(t, m.compressed)

	m1 = NewMessageWithRouter(newRouter)
	err = m1.Decode(mByts)
	if err != nil {
		t.Fatal(err)
	}
	myByts1, err = m1.Encode()
	if err != nil {
		t.Fatalf("err: %+v", err)
	}
	fmt.Println(myByts1)
	assert.Equal(t, mByts, myByts1)
	assert.False(t, m1.compressed)
	assert.Equal(t, m1, m)

}

func TestMessageKeepRouterCode(t *testing.T) {
	arr := []string{"A", "B", "C"}
	//arr := []string{"A"}

	newRouter := router.NewRouter()
	for i, v := range arr {
		//i += 10
		newRouter.AddRouteKV("1", v, uint16(i))
	}
	for _, v := range arr {
		m := NewMessage(Notify, 0, 0, v, nil, "", newRouter)
		mByts, err := m.Encode()
		if err != nil {
			t.Fatalf("err: %+v", err)
		}
		//fmt.Println(mByts)
		assert.True(t, m.compressed)

		m1 := Message{
			router: nil,
		}
		err = m1.Decode(mByts)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, m1.compressed)
		assert.Equal(t, m1.Route, "")
		assert.NotEqual(t, m1.Route, m.Route)

		myByts1, err := m1.Encode()
		if err != nil {
			t.Fatalf("err: %+v", err)
		}
		assert.Equal(t, myByts1, mByts)
	}

}

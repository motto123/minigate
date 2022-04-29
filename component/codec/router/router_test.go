package router

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRouter_AddRouteKV(t *testing.T) {
	exchangeName := "test"
	router := NewRouter()
	assert.True(t, router.AddRouteKV(exchangeName, "a", 1))
	assert.True(t, router.AddRouteKV(exchangeName, "b", 2))
	//fmt.Printf("%+v\n", router)
	assert.False(t, router.AddRouteKV(exchangeName, "a", 3))
	//fmt.Printf("%+v\n", router)
	assert.True(t, router.AddRouteKV(exchangeName, "d", 4))
	assert.True(t, router.AddRouteKV(exchangeName, "e", 5))

	assert.False(t, router.AddRouteKV("A", "e", 6))
	assert.True(t, router.AddRouteKV("B", "B", 7))

	name, _ := router.GetExchangeName("a")
	assert.Equal(t, name, "test")
	name, _ = router.GetExchangeName("e")
	assert.Equal(t, name, "test")
	name, _ = router.GetExchangeName("B")
	assert.Equal(t, name, "B")
	//t.Logf("%+v\n", router)
}

func TestRouter(t *testing.T) {
	router := NewRouter()
	n := 100000
	for i := 0; i < n; i++ {
		if router.Len() >= codeMax {
			return
		}
		router.AddRoute("A", fmt.Sprintf("%d", i))
	}
	name, _ := router.GetExchangeName("1")
	assert.Equal(t, name, "A")

	for i := 0; i < n; i++ {
		s := fmt.Sprintf("%d", i)
		code, _ := router.GetRouteCode(s)
		name, _ := router.GetRouteName(code)
		if name != s {
			t.Errorf("expect %s, but actual :%s\n", s, name)
		}
	}

}

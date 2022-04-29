package router

import (
	"fmt"
	"math/rand"
)

const (
	//maximum value of the uint16
	codeMax = 65536
)

type Router struct {
	//route:exchange
	exchanges map[string]string

	routes map[string]uint16
	codes  map[uint16]string
}

func NewRouter() *Router {
	return &Router{
		exchanges: make(map[string]string),
		routes:    make(map[string]uint16),
		codes:     make(map[uint16]string),
	}
}

// AddRouteKV 测试才能使用,正式环境不用使用
func (r *Router) AddRouteKV(exchangeName, routeName string, code uint16) bool {
	if r.Len() >= codeMax {
		fmt.Printf("AddRoute failed, becase routes len is large than codeMax %d\n", codeMax)
		return false
	}
	if code == 0 {
		fmt.Printf("AddRoute failed, becase code is not zero")
		return false
	}
	//var ok bool
	_, ok := r.routes[routeName]
	_, ok1 := r.codes[code]
	if !ok && !ok1 {
		r.routes[routeName] = code
		r.codes[code] = routeName
		if exchangeName != "" {
			r.exchanges[routeName] = exchangeName
		}
		return true
	}
	return false
}

func (r *Router) AddRoute(exchangeName, routeName string) bool {
	if r.Len() >= codeMax {
		fmt.Printf("AddRoute failed, becase routes len is large than codeMax %d\n", codeMax)
		return false
	}
	_, ok := r.routes[routeName]
	if ok {
		return true
	}
	ok = true
	n := rand.Intn(codeMax)
	for ok {
		code := uint16(n)
		if code == 0 {
			continue
		}
		_, ok = r.codes[code]
		if ok {
			n = rand.Intn(codeMax)
			continue
		}
		r.codes[code] = routeName
		r.routes[routeName] = code
		r.exchanges[routeName] = exchangeName
	}
	return true
}

func (r *Router) GetExchangeName(routeName string) (name string, ok bool) {
	name, ok = r.exchanges[routeName]
	return
}

func (r *Router) GetRouteName(code uint16) (name string, ok bool) {
	name, ok = r.codes[code]
	return
}

func (r *Router) GetRouteCode(name string) (code uint16, ok bool) {
	code, ok = r.routes[name]
	return
}

func (r Router) Len() int {
	return len(r.routes)
}

func (r *Router) GetRoutes() map[string]uint16 {
	return r.routes
}

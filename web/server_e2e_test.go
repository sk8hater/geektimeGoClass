// go:build e2e
package web

import (
	"fmt"
	"testing"
)

func TestServer(t *testing.T) {
	var h = NewHTTPServer()

	h.Get("/user", func(ctx *Context) {
		fmt.Println("first thing")
		fmt.Println("second thing")
	})

	h.Get("/order/detail", func(ctx *Context) {
		ctx.Resp.Write([]byte("HELLO WORLD"))
	})

	h.Get("/order/abc", func(ctx *Context) {
		ctx.Resp.Write([]byte("通配符匹配"))
	})

	// handler1 := func(ctx *Context) {
	// 	fmt.Println("处理第一件事")
	// }
	//
	// handler2 := func(ctx *Context) {
	// 	fmt.Println("处理第二件事")
	// }

	// 用户自己去管这种
	// h.addRoute(http.MethodGet, "/user", func(ctx *Context) {
	// 	handler1(ctx)
	// 	handler2(ctx)
	// })
	//
	// h.Get("/user", func(ctx *Context) {
	//
	// })

	// 用法一，完全委托给http包
	//http.ListenAndServe(":8081", h)

	// 用法二，自己手动管
	h.Start(":8081")
}

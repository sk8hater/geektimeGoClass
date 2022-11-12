package web

import (
	"net"
	"net/http"
)

type HandleFunc func(ctx *Context)

// 确保结构体实现了 Server 接口
var _ Server = &HttpServer{}

type Server interface {
	http.Handler
	Start(add string) error

	// AddRoute 增加路由注册的功能
	// method 是 HTTP 方法
	// path 是路由
	// handleFunc 是业务逻辑
	addRoute(method string, path string, handleFunc HandleFunc)
}

type HttpServer struct {
	*router
}

func NewHTTPServer() *HttpServer {
	return &HttpServer{
		router: newRouter(),
	}
}

func (h *HttpServer) Get(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodGet, path, handleFunc)
}

func (h *HttpServer) Post(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodPost, path, handleFunc)
}

// ServeHTTP 处理请求的入口
func (h *HttpServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	//  框架代码位置
	ctx := &Context{
		Req:  request,
		Resp: writer,
	}
	h.serve(ctx)
}

func (h *HttpServer) serve(ctx *Context) {
	// 查找路由，并且执行命中的业务逻辑
	info, ok := h.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || info.n.handler == nil {
		// 路由没有命中，返回404
		ctx.Resp.WriteHeader(404)
		ctx.Resp.Write([]byte("NOT FOUND"))
		return
	}
	ctx.pathParams = info.pathParams
	info.n.handler(ctx)
}

func (h *HttpServer) Start(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// 这里可以让用户注册 after start 回调
	// 比如往admin注册实例
	// 在这里执行一些业务需要的前置条件
	// （生命周期回调）

	return http.Serve(l, h)
}

func (h *HttpServer) Start1(addr string) error {
	return http.ListenAndServe(addr, h)
}

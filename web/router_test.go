package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func TestRouter_AddRoute(t *testing.T) {
	// 1. 构造路径树
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
		// 通配符匹配
		{
			method: http.MethodGet,
			path:   "/order/*",
		},
		{
			method: http.MethodGet,
			path:   "/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc/*",
		},
		// 参数路径
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/*",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail/:id",
		},
		{
			method: http.MethodPost,
			path:   "/:id",
		},
		// 正则匹配
		//{
		//	method: http.MethodDelete,
		//	path:   "/:id([0-9]+)",
		//},
		{
			method: http.MethodDelete,
			path:   "/reg/:id(.*)",
		},
		{
			method: http.MethodDelete,
			path:   "/:name(^.+$)/abc",
		},
	}

	var mockHandler HandleFunc = func(ctx *Context) {}

	// 2. 验证路径树
	r := newRouter()

	for _, route := range testRoutes {
		r.addRoute(route.method, route.path, mockHandler)
	}

	// 3. 断言
	// HandleFunc 不能用assert
	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: &node{
				path:    "/",
				handler: mockHandler,
				children: map[string]*node{
					"user": &node{
						path:    "user",
						handler: mockHandler,
						children: map[string]*node{
							"home": &node{
								path:     "home",
								children: nil,
								handler:  mockHandler,
							},
						},
					},
					"order": &node{
						path: "order",
						children: map[string]*node{
							"detail": &node{
								path:     "detail",
								children: nil,
								handler:  mockHandler,
								paramChild: &node{
									path:    ":id",
									handler: mockHandler,
								},
							},
						},
						starChild: &node{
							path:    "*",
							handler: mockHandler,
						},
					},
					"param": {
						path: "param",
						paramChild: &node{
							path: ":id",
							starChild: &node{
								path:    "*",
								handler: mockHandler,
							},
							children: map[string]*node{"detail": {path: "detail", handler: mockHandler}},
							handler:  mockHandler,
						},
					},
				},
				starChild: &node{
					path:    "*",
					handler: mockHandler,
					children: map[string]*node{
						"abc": &node{
							path: "abc",
							starChild: &node{
								path:    "*",
								handler: mockHandler,
							},
							handler: mockHandler,
						},
					},
					starChild: &node{
						path:    "*",
						handler: mockHandler,
					},
				},
			},
			http.MethodPost: &node{
				path:    "/",
				handler: mockHandler,
				children: map[string]*node{
					"login": &node{
						path:     "login",
						children: nil,
						handler:  mockHandler,
					},
					"order": &node{
						path: "order",
						children: map[string]*node{
							"create": &node{
								path:     "create",
								children: nil,
								handler:  mockHandler,
							},
						},
					},
				},
				paramChild: &node{
					path:    ":id",
					handler: mockHandler,
				},
			},
			http.MethodDelete: &node{
				path: "/",
				children: map[string]*node{
					"reg": {
						path: "reg",
						regChild: &node{
							path:    ":id(.*)",
							handler: mockHandler,
						},
					},
				},
				regChild: &node{
					path: ":name(^.+$)",
					children: map[string]*node{
						"abc": {
							path:    "abc",
							handler: mockHandler,
						},
					},
				},
			},
		},
	}

	msg, ok := r.equal(wantRouter)
	assert.True(t, ok, msg)

	r = newRouter()
	assert.PanicsWithValue(t, "web：路径不能为空字符串", func() {
		r.addRoute(http.MethodGet, "", mockHandler)
	})

	r = newRouter()
	assert.PanicsWithValue(t, "web：路径必须以 / 开头", func() {
		r.addRoute(http.MethodGet, "user", mockHandler)
	})

	r = newRouter()
	assert.PanicsWithValue(t, "web：路径不能以 / 结尾", func() {
		r.addRoute(http.MethodGet, "/user/root/", mockHandler)
	})

	r = newRouter()
	assert.PanicsWithValue(t, "web：路径不能出现连续的 /", func() {
		r.addRoute(http.MethodGet, "/user//root", mockHandler)
	})

	r = newRouter()
	assert.PanicsWithValue(t, "web：路径不能出现连续的 /", func() {
		r.addRoute(http.MethodGet, "//a/b", mockHandler)
	})

	r = newRouter()
	r.addRoute(http.MethodGet, "/", mockHandler)
	assert.PanicsWithValue(t, "web: 路径冲突，重复注册[/]", func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	})
	r = newRouter()
	r.addRoute(http.MethodGet, "/a", mockHandler)
	assert.PanicsWithValue(t, "web: 路径冲突，重复注册[/a]", func() {
		r.addRoute(http.MethodGet, "/a", mockHandler)
	})
	r = newRouter()
	r.addRoute(http.MethodGet, "/*", mockHandler)
	assert.PanicsWithValue(t, "web: 路径冲突，重复注册[/*]", func() {
		r.addRoute(http.MethodGet, "/*", mockHandler)
	})
	r = newRouter()
	r.addRoute(http.MethodGet, "/:id", mockHandler)
	assert.PanicsWithValue(t, "web: 路径冲突，重复注册[/:id]", func() {
		r.addRoute(http.MethodGet, "/:id", mockHandler)
	})
	r = newRouter()
	r.addRoute(http.MethodGet, "/:id(.*)", mockHandler)
	assert.PanicsWithValue(t, "web: 路径冲突，重复注册[/:id(.*)]", func() {
		r.addRoute(http.MethodGet, "/:id(.*)", mockHandler)
	})

	r = newRouter()
	r.addRoute(http.MethodGet, "/a/*", mockHandler)
	assert.PanicsWithValue(t, "web：不允许同时注册路径参数，通配符路径或正则路径，已有通配符路径", func() {
		r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	})

	r = newRouter()
	r.addRoute(http.MethodGet, "/a/*", mockHandler)
	assert.PanicsWithValue(t, "web：不允许同时注册路径参数，通配符路径或正则路径，已有通配符路径", func() {
		r.addRoute(http.MethodGet, "/a/:id(.*)", mockHandler)
	})

	r = newRouter()
	r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	assert.PanicsWithValue(t, "web：不允许同时注册路径参数，通配符路径或正则路径，已有参数路径", func() {
		r.addRoute(http.MethodGet, "/a/*", mockHandler)
	})

	r = newRouter()
	r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	assert.PanicsWithValue(t, "web：不允许同时注册路径参数，通配符路径或正则路径，已有参数路径", func() {
		r.addRoute(http.MethodGet, "/a/:id(.*)", mockHandler)
	})

	r = newRouter()
	r.addRoute(http.MethodGet, "/a/:id(.*)", mockHandler)
	assert.PanicsWithValue(t, "web：不允许同时注册路径参数，通配符路径或正则路径，已有正则路径", func() {
		r.addRoute(http.MethodGet, "/a/*", mockHandler)
	})

	r = newRouter()
	r.addRoute(http.MethodGet, "/a/:id(.*)", mockHandler)
	assert.PanicsWithValue(t, "web：不允许同时注册路径参数，通配符路径或正则路径，已有正则路径", func() {
		r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	})

	r = newRouter()
	r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	assert.PanicsWithValue(t, "web: 路径冲突，已注册[:id]，重复注册[:age]", func() {
		r.addRoute(http.MethodGet, "/a/:age/abc", mockHandler)
	})

	r = newRouter()
	r.addRoute(http.MethodGet, "/a/:id(.*)", mockHandler)
	assert.PanicsWithValue(t, "web: 路径冲突，已注册[:id(.*)]，重复注册[:age(.*)]", func() {
		r.addRoute(http.MethodGet, "/a/:age(.*)/abc", mockHandler)
	})

	r = newRouter()
	assert.PanicsWithValue(t, "regexp: Compile(`.(.*`): error parsing regexp: missing closing ): `.(.*`", func() {
		r.addRoute(http.MethodGet, "/a/:age(.(.*)", mockHandler)
	})
}

// string 返回错误信息，帮助定位问题
// bool 是否相等
func (r *router) equal(y *router) (string, bool) {
	// 非空判断
	if r == nil || y == nil {
		return fmt.Sprintf("空的路径树"), false
	}

	// 便利
	for k, v := range r.trees {
		dst, ok := y.trees[k]
		if !ok {
			return fmt.Sprintf("找不到对应的 http method"), false
		}

		msg, equal := v.equal(dst)
		if !equal {
			return msg, false
		}

	}
	return "", true
}

func (n *node) equal(y *node) (string, bool) {
	// 非空判断
	if n == nil || y == nil {
		return fmt.Sprintf("空节点"), false
	}

	if n.path != y.path {
		return fmt.Sprintf("节点路径不匹配"), false
	}

	if len(n.children) != len(y.children) {
		return fmt.Sprintf("子节点数量不相等"), false
	}

	// 比较 handler
	nHandler := reflect.ValueOf(n.handler)
	yHandler := reflect.ValueOf(y.handler)
	if nHandler != yHandler {
		return fmt.Sprintf("handler 不相等"), false
	}

	// 比较 starChild
	if n.starChild != nil || y.starChild != nil {
		msg, equal := n.starChild.equal(y.starChild)
		if !equal {
			return msg, false
		}
	}

	// 比较 paramChild
	if n.paramChild != nil || y.paramChild != nil {
		msg, equal := n.paramChild.equal(y.paramChild)
		if !equal {
			return msg, false
		}
	}

	// 比较 regChild
	if n.regChild != nil || y.regChild != nil {
		msg, equal := n.regChild.equal(y.regChild)
		if !equal {
			return msg, false
		}
	}

	for path, c := range n.children {
		dst, ok := y.children[path]
		if !ok {
			return fmt.Sprintf("子节点 %s 不存在", path), false
		}
		msg, equal := c.equal(dst)
		if !equal {
			return msg, false
		}
	}
	return "", true
}

func TestRouter_FindRoute(t *testing.T) {
	// 1. 构造路径树
	testRoutes := []struct {
		method string
		path   string
	}{
		// 根节点
		{
			method: http.MethodDelete,
			path:   "/",
		},
		// 单节点
		{
			method: http.MethodPut,
			path:   "/login",
		},
		// 双节点
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		// 通配符匹配
		{
			method: http.MethodGet,
			path:   "/star/*",
		},
		{
			method: http.MethodGet,
			path:   "/star/*/abc",
		},
		{
			method: http.MethodGet,
			path:   "/star/a/*",
		},
		{
			method: http.MethodGet,
			path:   "/star/a/b/c",
		},
		// 参数路径匹配
		{
			method: http.MethodPost,
			path:   "/params/:username",
		},
		{
			method: http.MethodPost,
			path:   "/params/:username/detail",
		},
		{
			method: http.MethodPost,
			path:   "/params/:username/*",
		},
		// 正则
		{
			method: http.MethodPost,
			path:   "/reg/:id(.*)",
		},
		{
			method: http.MethodPost,
			path:   "/:id([0-9]+)/home",
		},
	}

	var mockHandler HandleFunc = func(ctx *Context) {}
	var mockHandlerClass HandleFunc = func(ctx *Context) {}

	r := newRouter()
	for _, route := range testRoutes {
		if route.path == "/star/a/*" {
			r.addRoute(route.method, route.path, mockHandlerClass)
		} else {
			r.addRoute(route.method, route.path, mockHandler)
		}
	}

	testCases := []struct {
		name string

		method string
		path   string

		wantFound bool
		wantInfo  *matchInfo
	}{
		{
			// 方法不存在
			name:   "method not found",
			method: http.MethodOptions,
			path:   "/",
		},
		{
			// 根节点
			name:      "root",
			method:    http.MethodDelete,
			path:      "/",
			wantFound: true,
			wantInfo: &matchInfo{
				n: &node{
					path:     "/",
					handler:  mockHandler,
					children: nil,
				},
			},
		},
		{
			// 单节点
			name:      "login",
			method:    http.MethodPut,
			path:      "/login",
			wantFound: true,
			wantInfo: &matchInfo{
				n: &node{
					path:    "login",
					handler: mockHandler,
				},
			},
		},
		{
			// 双节点
			name:      "order detail",
			method:    http.MethodGet,
			path:      "/order/detail",
			wantFound: true,
			wantInfo: &matchInfo{
				n: &node{
					path:     "detail",
					handler:  mockHandler,
					children: nil,
				},
			},
		},
		{
			// 命中了 但是没有 handler
			name:      "order",
			method:    http.MethodGet,
			path:      "/order",
			wantFound: true,
			wantInfo: &matchInfo{
				n: &node{
					path:    "order",
					handler: nil,
					children: map[string]*node{
						"detail": &node{
							path:     "detail",
							handler:  mockHandler,
							children: nil,
						},
					},
				},
			},
		},
		{
			// 命中通配符路径 /star/*
			name:      "star abc",
			method:    http.MethodGet,
			path:      "/star/abc",
			wantFound: true,
			wantInfo: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
					children: map[string]*node{
						"abc": &node{
							path:    "abc",
							handler: mockHandler,
						},
					},
				},
			},
		},
		{
			// 命中通配符路径 /star/*/abc
			name:      "star abc",
			method:    http.MethodGet,
			path:      "/star/abc/abc",
			wantFound: true,
			wantInfo: &matchInfo{
				n: &node{
					path:    "abc",
					handler: mockHandler,
				},
			},
		},
		{
			// 通配符匹配: 结尾匹配多段路径
			name:      "star a b c d",
			method:    http.MethodGet,
			path:      "/star/a/b/c/d",
			wantFound: true,
			wantInfo: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandlerClass,
				},
			},
		},
		{
			// 参数路径: /params/:username
			name:      "params username",
			method:    http.MethodPost,
			path:      "/params/why",
			wantFound: true,
			wantInfo: &matchInfo{
				n: &node{
					path:    ":username",
					handler: mockHandler,
					children: map[string]*node{
						"detail": &node{
							path:    "detail",
							handler: mockHandler,
						},
					},
					starChild: &node{
						path:    "*",
						handler: mockHandler,
					},
				},
				pathParams: map[string]string{
					"username": "why",
				},
			},
		},
		{
			// 参数路径: /params/:username/detail
			name:      "params username detail",
			method:    http.MethodPost,
			path:      "/params/why/detail",
			wantFound: true,
			wantInfo: &matchInfo{
				n: &node{
					path:    "detail",
					handler: mockHandler,
				},
				pathParams: map[string]string{
					"username": "why",
				},
			},
		},
		{
			// 参数路径: /params/:username/*
			name:      "params username abc",
			method:    http.MethodPost,
			path:      "/params/why/abc",
			wantFound: true,
			wantInfo: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
				pathParams: map[string]string{
					"username": "why",
				},
			},
		},
		{
			// 正则路径: /reg/:id(.*)
			name:      "reg id any",
			method:    http.MethodPost,
			path:      "/reg/why",
			wantFound: true,
			wantInfo: &matchInfo{
				n: &node{
					path:    ":id(.*)",
					handler: mockHandler,
				},
				pathParams: map[string]string{
					"id": "why",
				},
			},
		},
		{
			// 正则路径: /:id([0-9]+)/home
			name:      "reg id number",
			method:    http.MethodPost,
			path:      "/123/home",
			wantFound: true,
			wantInfo: &matchInfo{
				n: &node{
					path:    "home",
					handler: mockHandler,
				},
				pathParams: map[string]string{
					"id": "123",
				},
			},
		},
		{
			// 正则路径: /:id([0-9]+)/home，没有匹配上
			name:      "reg id number",
			method:    http.MethodPost,
			path:      "/why/home",
			wantFound: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.wantFound, found)
			if !found {
				return
			}
			assert.Equal(t, tc.wantInfo.pathParams, info.pathParams)
			msg, ok := info.n.equal(tc.wantInfo.n)
			assert.True(t, ok, msg)
		})
	}

}

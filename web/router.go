package web

import (
	"fmt"
	"regexp"
	"strings"
)

// 用来支持对路径树的操作
// 代表路径树（森林）
type router struct {
	// key: HTTP method =》根节点
	// value: 子节点 =》 path
	trees map[string]*node
}

func newRouter() *router {
	return &router{
		trees: map[string]*node{},
	}
}

// addRoute 添加限制：
// path 必须以 / 开头，不能以 / 结尾 且不能出现连续的 //
func (r *router) addRoute(method string, path string, handleFunc HandleFunc) {
	if path == "" {
		panic("web：路径不能为空字符串")
	}

	root, ok := r.trees[method]
	if !ok {
		// 还没有根节点
		root = &node{
			path:     "/",
			children: nil,
			handler:  nil,
		}
		r.trees[method] = root
	}

	if path[0] != '/' {
		panic("web：路径必须以 / 开头")
	}

	// 根节点特殊处理
	if path == "/" {
		if root.handler != nil {
			panic("web: 路径冲突，重复注册[/]")
		}
		root.handler = handleFunc
		return
	}

	if path[len(path)-1] == '/' {
		panic("web：路径不能以 / 结尾")
	}

	// 切割 path
	segs := strings.Split(path, "/")
	segs = segs[1:]
	for _, seg := range segs {
		if seg == "" {
			panic("web：路径不能出现连续的 /")
		}
		// 递归找children
		// 不存在就创建
		root = root.childOrCreate(seg)
	}
	if root.handler != nil {
		panic(fmt.Sprintf("web: 路径冲突，重复注册[%s]", path))
	}
	root.handler = handleFunc
}

func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	// 树的深度遍历查找
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}
	if path == "/" {
		return &matchInfo{
			n: root,
		}, true
	}

	path = strings.Trim(path, "/")
	segs := strings.Split(path, "/")
	var (
		pathParams   map[string]string
		starNodeTemp *node
	)
	for _, seg := range segs {
		child, starNode, isRegChild, isParamChild, found := root.childOf(seg)

		if !found && starNodeTemp != nil {
			root = starNodeTemp
			break
		}
		if !found {
			return nil, false
		}
		if starNode != nil {
			starNodeTemp = starNode
		}
		root = child

		if isRegChild {
			matchRegs := regexp.MustCompile(`:(.*?)\((.*)\)`)
			regs := matchRegs.FindStringSubmatch(child.path)
			// 注册时已校验,此处不校验
			reg := regexp.MustCompile(regs[2])
			matched := reg.MatchString(seg)
			if !matched && starNodeTemp != nil {
				root = starNodeTemp
				break
			}
			if !matched {
				return nil, false
			}
			if pathParams == nil {
				pathParams = make(map[string]string)
			}
			pathParams[regs[1]] = seg
		}

		if isParamChild {
			if pathParams == nil {
				pathParams = make(map[string]string)
			}
			pathParams[child.path[1:]] = seg
		}
	}

	return &matchInfo{
		n:          root,
		pathParams: pathParams,
	}, true
}

type node struct {
	// 相当于 key
	path string

	// 子 path 到子节点的映射
	children map[string]*node

	// 通配符匹配
	starChild *node

	// 路径参数
	paramChild *node

	// 正则匹配
	regChild *node

	// 代表用户注册的业务逻辑
	handler HandleFunc
}

func (n *node) childOrCreate(path string) *node {
	matchRegs := regexp.MustCompile(`:(.*?)\((.*)\)`)
	regs := matchRegs.FindStringSubmatch(path)
	if regs != nil {
		if n.starChild != nil {
			panic("web：不允许同时注册路径参数，通配符路径或正则路径，已有通配符路径")
		}
		if n.paramChild != nil {
			panic("web：不允许同时注册路径参数，通配符路径或正则路径，已有参数路径")
		}
		if n.regChild != nil && n.regChild.path != path {
			panic(fmt.Sprintf("web: 路径冲突，已注册[%s]，重复注册[%s]", n.regChild.path, path))
		}
		// 校验正则是否合法
		regexp.MustCompile(regs[2])
		if n.regChild == nil {
			n.regChild = &node{
				path: path,
			}
		}
		return n.regChild
	}

	if path[0] == ':' {
		if n.starChild != nil {
			panic("web：不允许同时注册路径参数，通配符路径或正则路径，已有通配符路径")
		}
		if n.regChild != nil {
			panic("web：不允许同时注册路径参数，通配符路径或正则路径，已有正则路径")
		}

		if n.paramChild != nil && n.paramChild.path != path {
			panic(fmt.Sprintf("web: 路径冲突，已注册[%s]，重复注册[%s]", n.paramChild.path, path))
		}
		if n.paramChild == nil {
			n.paramChild = &node{
				path: path,
			}
		}
		return n.paramChild
	}

	if path == "*" {
		if n.paramChild != nil {
			panic("web：不允许同时注册路径参数，通配符路径或正则路径，已有参数路径")
		}
		if n.regChild != nil {
			panic("web：不允许同时注册路径参数，通配符路径或正则路径，已有正则路径")
		}
		if n.starChild == nil {
			n.starChild = &node{
				path: path,
			}
		}
		return n.starChild
	}
	if n.children == nil {
		n.children = make(map[string]*node)
	}
	child, ok := n.children[path]
	if !ok {
		// 新建
		child = &node{
			path: path,
		}
		n.children[path] = child
	}
	return child
}

// childOf 优先考虑静态查找，其次是参数路径，匹配不成功考虑通配符查找
// 第一个返回值是路径节点
// 第二个返回值标记是否已匹配到通配符
// 第三个返回值标记是否是正则匹配
// 第四个返回值标记是否是参数路径
// 第五个返回值标记是否找到
func (n *node) childOf(path string) (*node, *node, bool, bool, bool) {
	if n.children == nil && n.regChild != nil {
		return n.regChild, n.starChild, true, false, true
	}
	if n.children == nil && n.paramChild != nil {
		return n.paramChild, n.starChild, false, true, true
	}
	if n.children == nil {
		return n.starChild, n.starChild, false, false, n.starChild != nil
	}
	child, ok := n.children[path]
	if !ok && n.regChild != nil {
		return n.regChild, n.starChild, true, false, true
	}
	if !ok && n.paramChild != nil {
		return n.paramChild, n.starChild, false, true, true
	}
	if !ok {
		return n.starChild, n.starChild, false, false, n.starChild != nil
	}
	return child, n.starChild, false, false, ok
}

type matchInfo struct {
	n          *node
	pathParams map[string]string
}

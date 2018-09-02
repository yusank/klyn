// Copyright 2018 Yusan Kurban. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package klyn

import (
	"reflect"
	"runtime"
)

const (
	Version                = "v0.0.1"
	defaultMultipartMemory = 32 << 10 // 16M
)

type HandlerFunc func(*Context)
type HandlersChain []HandlerFunc

// Last returns the last handler in the chain. ie. the last handler is the main own.
func (c HandlersChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

type RouteInfo struct {
	Method  string
	Path    string
	Handler string
}

type RoutesInfo []RouteInfo

// Core core of framework
type Core struct {
	RouterGroup

	trees methodTrees
}

var _ KRouter = &Core{}

func New() *Core {
	core := &Core{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "",
			root:     false,
		},

		trees: make(methodTrees, 0, 9),
	}

	core.RouterGroup.core = core

	return core
}

func Default() *Core {
	core := New()
	return core
}

func (core *Core) UseMiddleware(middleware ...HandlerFunc) KRoutes {
	core.RouterGroup.UseMiddleware(middleware...)
	return core
}

func (core *Core) addRouter(method, path string, handlers HandlersChain) {
	assert1(path[0] == '/', "path must begin with '/'")
	assert1(method != "", "HTTP method can not be empty")
	assert1(len(handlers) > 0, "there must be at least one handler")

	root := core.trees.get(method)
	if root == nil {
		root = new(node)
		core.trees = append(core.trees, methodTree{method: method, root: root})
	}

	root.addRoute(path, handlers)
}

func (core *Core) Routes() (routes RoutesInfo) {
	for _, tree := range core.trees {
		iterate(tree.method, "", routes, tree.root)
	}

	return routes
}

func iterate(method, path string, routes RoutesInfo, root *node) RoutesInfo {
	path += root.path
	if len(root.handlers) > 0 {
		routes = append(routes, RouteInfo{
			Method:  method,
			Path:    path,
			Handler: nameOfFunction(root.handlers.Last()),
		})
	}
	for _, child := range root.children {
		routes = iterate(path, method, routes, child)
	}
	return routes
}

func nameOfFunction(handler interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
}

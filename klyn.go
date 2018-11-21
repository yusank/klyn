// Copyright 2018 Yusan Kurban. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package klyn

import (
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"sync"
)

const (
	// Version - version of project
	Version                = "v0.0.1"
	defaultMultipartMemory = 32 << 10 // 16M
)

type HandlerFunc func(*Context)
type HandlersChain []HandlerFunc
type K map[string]interface{}

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

	UseRawPath             bool
	UnescapePathValues     bool
	HandleMethodNotAllowed bool

	ForwardByClientIP bool

	trees methodTrees
	pool  sync.Pool
}

var _ KRouter = &Core{}
var (
	default405Body = []byte("405 method not allowed")
	default404Body = []byte("404 page not found")
)

func New() *Core {
	core := &Core{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "",
			root:     true,
		},

		HandleMethodNotAllowed: true,
		trees: make(methodTrees, 0, 9),
	}
	core.pool.New = func() interface{} {
		return core.allocateContext()
	}

	core.RouterGroup.core = core

	return core
}

func Default() *Core {
	core := New()
	core.UseMiddleware(Logger())
	return core
}

func (core *Core) allocateContext() *Context {
	return &Context{core: core}
}

func (core *Core) UseMiddleware(middleware ...HandlerFunc) KRoutes {
	core.RouterGroup.UseMiddleware(middleware...)
	return core
}

func (core *Core) addRouter(method, path string, handlers HandlersChain) {
	assert1(path[0] == '/', "path must begin with '/'")
	assert1(method != "", "HTTP method can not be empty")
	assert1(len(handlers) > 0, "there must be at least one handler")

	printRouter(method, path, handlers)
	root := core.trees.get(method)
	if root == nil {
		root = new(node)
		core.trees = append(core.trees, methodTree{method: method, root: root})
	}

	root.addRoute(path, handlers)
}

func printRouter(method, path string, handlers HandlersChain) {
	handlerName := nameOfFunction(handlers.Last())
	log.Printf("%-7s  %-20s --> %s (handlers:%d) \n", method, path, handlerName, len(handlers))
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

func (core *Core) Service(addr ...string) (err error) {
	address := resolveAddress(addr)
	log.Println("start service on:", address)
	err = http.ListenAndServe(address, core)
	return
}

func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); port != "" {
			return ":" + port
		}
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too much parameters")
	}
}

// ServeHTTP conforms to the http.Handler interface.
func (core *Core) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := core.pool.Get().(*Context)
	c.memWriter.reset(w)
	c.Request = req
	c.reset()

	core.handleHTTPRequest(c)

	core.pool.Put(c)
}

func (core *Core) handleHTTPRequest(c *Context) {
	method := c.Request.Method
	path := c.Request.URL.Path
	unescape := false
	if core.UseRawPath && len(c.Request.URL.RawPath) > 0 {
		path = c.Request.URL.RawPath
		unescape = core.UnescapePathValues
	}

	t := core.trees
	for i, tl := 0, len(t); i < tl; i++ {
		if t[i].method == method {
			root := t[i].root
			handlers, params, _ := root.getValue(path, c.Params, unescape)
			if handlers != nil {
				c.handlers = handlers
				c.Params = params
				c.Next()
				return
			}
		}
	}

	if core.HandleMethodNotAllowed {
		for _, tree := range core.trees {
			if tree.method == method {
				continue
			}

			handlers, _, _ := tree.root.getValue(path, c.Params, unescape)
			if handlers != nil {
				serverError(c, http.StatusMethodNotAllowed, default405Body)
				return
			}
		}
	}

	serverError(c, http.StatusNotFound, default404Body)
	return
}

func serverError(c *Context, code int, defaultMessage []byte) {
	c.memWriter.status = code
	c.Next()
	if !c.memWriter.Written() {
		if c.memWriter.Status() == code {
			c.memWriter.Header()["Content-Type"] = []string{"text/plain"}
			c.Writer.Write(defaultMessage)
		} else {
			c.memWriter.WriteHeaderNow()
		}
	}
}

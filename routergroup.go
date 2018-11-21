// Copyright 2018 Yusan Kurban. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package klyn

import (
	"regexp"
)

type KRouter interface {
	KRoutes
	Group(string, ...HandlerFunc) *RouterGroup
}

type KRoutes interface {
	UseMiddleware(...HandlerFunc) KRoutes

	Handle(string, string, ...HandlerFunc) KRoutes
	Any(string, ...HandlerFunc) KRoutes
	GET(string, ...HandlerFunc) KRoutes
	POST(string, ...HandlerFunc) KRoutes
	PUT(string, ...HandlerFunc) KRoutes
	DELETE(string, ...HandlerFunc) KRoutes
	PATCH(string, ...HandlerFunc) KRoutes
	OPTIONS(string, ...HandlerFunc) KRoutes
	HEAD(string, ...HandlerFunc) KRoutes
}

type RouterGroup struct {
	Handlers HandlersChain
	basePath string
	core     *Core
	root     bool
}

func (rg *RouterGroup) handle(method, relativePath string, handlers HandlersChain) KRoutes {
	absolutePath := rg.calculatePath(relativePath)
	handlers = rg.combineHandlers(handlers)
	rg.core.addRouter(method, absolutePath, handlers)
	return rg.returnObj()
}

func (rg *RouterGroup) UseMiddleware(middleware ...HandlerFunc) KRoutes {
	rg.Handlers = append(rg.Handlers, middleware...)

	return rg.returnObj()
}

func (rg *RouterGroup) Group(relativePath string, handler ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		Handlers: rg.combineHandlers(handler),
		basePath: rg.calculatePath(relativePath),
		core:     rg.core,
	}
}

func (rg *RouterGroup) Handle(method, relativePath string, handlers ...HandlerFunc) KRoutes {
	if matches, err := regexp.MatchString("^[A-Z]+$", method); !matches || err != nil {
		panic("http method " + method + " is not valid")
	}

	return rg.handle(method, relativePath, handlers)
}

func (rg *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) KRoutes {
	return rg.handle("GET", relativePath, handlers)
}

func (rg *RouterGroup) POST(relativePath string, handlers ...HandlerFunc) KRoutes {
	return rg.handle("POST", relativePath, handlers)
}

func (rg *RouterGroup) PUT(relativePath string, handlers ...HandlerFunc) KRoutes {
	return rg.handle("PUT", relativePath, handlers)
}

func (rg *RouterGroup) DELETE(relativePath string, handlers ...HandlerFunc) KRoutes {
	return rg.handle("DELETE", relativePath, handlers)
}

func (rg *RouterGroup) PATCH(relativePath string, handlers ...HandlerFunc) KRoutes {
	return rg.handle("PATCH", relativePath, handlers)
}

func (rg *RouterGroup) OPTIONS(relativePath string, handlers ...HandlerFunc) KRoutes {
	return rg.handle("OPTIONS", relativePath, handlers)
}

func (rg *RouterGroup) HEAD(relativePath string, handlers ...HandlerFunc) KRoutes {
	return rg.handle("HEAD", relativePath, handlers)
}

// Any - register all method
func (rg *RouterGroup) Any(relativePath string, handlers ...HandlerFunc) KRoutes {
	rg.handle("GET", relativePath, handlers)
	rg.handle("POST", relativePath, handlers)
	rg.handle("PUT", relativePath, handlers)
	rg.handle("DELETE", relativePath, handlers)
	rg.handle("PATCH", relativePath, handlers)
	rg.handle("OPTIONS", relativePath, handlers)
	rg.handle("HEAD", relativePath, handlers)
	rg.handle("CONNECT", relativePath, handlers)
	rg.handle("TRACE", relativePath, handlers)

	return rg.returnObj()
}

func (rg *RouterGroup) combineHandlers(handlers HandlersChain) HandlersChain {
	finalSize := len(rg.Handlers) + len(handlers)
	if finalSize >= int(abortIndex) {
		panic("too many handlers")
	}
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, rg.Handlers)
	copy(mergedHandlers[len(rg.Handlers):], handlers)
	return mergedHandlers
}

func (rg *RouterGroup) returnObj() KRoutes {
	if rg.root {
		return rg.core
	}
	return rg
}

func (rg *RouterGroup) calculatePath(relativePath string) string {
	return joinPaths(rg.basePath, relativePath)
}

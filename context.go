// Copyright 2018 Yusan Kurban. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package klyn

import (
	"math"
	"net/http"

	"encoding/json"
)

const (
	abortIndex  int8 = math.MaxInt8 / 2
	jsonContent      = "application/json; charset=utf-8"
)

type Context struct {
	memWriter responseWriter
	Request   *http.Request
	Writer    ResponseWriter
	Params    Params

	handlers  HandlersChain
	core      *Core
	index     int8
	cachePool map[string]interface{} // temp pool for context
}

// reset context
func (c *Context) reset() {
	c.Writer = &c.memWriter
	c.handlers = nil
	c.index = -1
	c.cachePool = nil
}

// Copy returns a copy of the current context that can be safely used outside the request's scope.
// This has to be used when the context has to be passed to a goroutine.
func (c *Context) Copy() *Context {
	var cp = *c
	cp.memWriter.ResponseWriter = nil
	cp.Writer = &cp.memWriter
	cp.index = abortIndex
	cp.handlers = nil
	return &cp
}

func (c *Context) Handler() HandlerFunc {
	return c.handlers.Last()
}

func (c *Context) HandlerName() string {
	return nameOfFunction(c.handlers.Last())
}

func (c *Context) Next() {
	c.index++
	for s := int8(len(c.handlers)); c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) IsAbort() bool {
	return c.index < abortIndex
}

func (c *Context) Abort() {
	c.index = abortIndex
}

// JSON - write response as json type
func (c *Context) JSON(code int, v interface{}) {
	c.Status(code)
	c.jsonContent()
	body, _ := json.Marshal(v)
	c.Writer.Write(body)
	return
}

// Data - write response as contentType
func (c *Context) Data(code int, contentType string, data []byte) {
	c.Status(code)
	c.Writer.Header().Set("Content-Type", contentType)
	c.Writer.Write(data)
}

func (c *Context) AbortWithStatus(code int) {
	c.Abort()
	c.Writer.WriteHeader(code)
	return
}

func (c *Context) AbortWithJSON(code int, v interface{}) {
	c.Abort()
	c.JSON(code, v)
}

// Status sets the HTTP response code.
func (c *Context) Status(code int) {
	c.memWriter.WriteHeader(code)
}

// write json content type to header
func (c *Context) jsonContent() {
	c.Writer.Header().Set("Content-Type", jsonContent)
}

// Set - set key-value to Context.cachePool
func (c *Context) Set(key string, value interface{}) {
	if c.cachePool == nil {
		c.cachePool = make(map[string]interface{})
	}

	c.cachePool[key] = value
}

// Get - get value from Context.cachePool
func (c *Context) Get(key string) (value interface{}, exist bool) {
	value, exist = c.cachePool[key]
	return
}

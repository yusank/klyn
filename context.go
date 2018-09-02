// Copyright 2018 Yusan Kurban. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package klyn

import (
	"encoding/json"
	"net/http"
)

type Context struct {
	Request *http.Request
	Writer  http.ResponseWriter
}

func (c *Context) AbortWithStatus(code int) {
	c.Writer.WriteHeader(code)
	return
}
func (c *Context) JSON(v interface{}) {
	marshal, _ := json.Marshal(v)
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write(marshal)
	return
}

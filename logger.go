package klyn

import (
	"fmt"
	"net/http"
	"time"

	"git.yusank.cn/yusank/klyn-log"
)

var (
	defaultKlynLog = klynlog.DefaultLogger()
)

// Logger log middleware
func Logger() HandlerFunc {
	return LoggerWithWriter()
}

// LoggerWithWriter -
func LoggerWithWriter(except ...string) HandlerFunc {

	var skip map[string]struct{}

	if l := len(except); l > 0 {
		skip = make(map[string]struct{}, l)

		for _, path := range except {
			skip[path] = struct{}{}
		}
	}

	return func(c *Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// handle request
		c.Next()

		// not in except list
		if _, ok := skip[path]; !ok {
			end := time.Now()
			useTime := end.Sub(start)
			clientIP := c.ClientIP()
			method := c.Request.Method
			statusCode := c.Writer.Status()

			if raw != "" {
				path += "?" + raw
			}

			//lFunc := logFuncForStatus(statusCode)
			defaultKlynLog.Info(map[string]interface{}{
				"clientIP":   clientIP,
				"method":     method,
				"path":       path,
				"statusCode": statusCode,
				"time":       end.Format("2006/01/02 15:04:05"),
				"useTime":    fmt.Sprintf("%v", useTime),
			})

		}
	}
}

func logFuncForStatus(code int) klynlog.LogFunc {
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return defaultKlynLog.Info
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return defaultKlynLog.Warn
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return defaultKlynLog.Error
	default:
		return defaultKlynLog.Debug
	}
}

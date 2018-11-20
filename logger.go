package klyn

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	DefaultWriter      io.Writer = os.Stdout
	DefaultErrorWriter io.Writer = os.Stderr

	green       = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white       = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow      = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red         = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue        = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta     = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan        = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset       = string([]byte{27, 91, 48, 109})
	useColorLog = true
)

func DisableColorLog() {
	useColorLog = false
}

// Logger log middleware
func Logger() HandlerFunc {
	return LoggerWithWriter(DefaultWriter)
}

// LoggerWithWriter -
func LoggerWithWriter(out io.Writer, except ...string) HandlerFunc {
	isTerm := true
	if _, ok := out.(*os.File); !ok || (os.Getenv("TERN") == "dumb") || !useColorLog {
		isTerm = false
	}

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
			var statusColor, methodColor, resetColor string
			if isTerm {
				statusColor = colorForStatus(statusCode)
				methodColor = colorForMethod(method)
				resetColor = reset
			}

			if raw != "" {
				path += "?" + raw
			}

			fmt.Fprintf(out, "[KLYN] %v |%s %3d %s| %13v | %15s |%s %-7s %s| %s\n",
				end.Format("2006/01/02 - 15:04:05"),
				statusColor, statusCode, resetColor,
				useTime,
				clientIP,
				methodColor, method, resetColor,
				path,
			)
		}
	}
}

func colorForStatus(code int) string {
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return yellow
	default:
		return red
	}
}

func colorForMethod(method string) string {
	switch method {
	case "GET":
		return blue
	case "POST":
		return cyan
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return green
	case "HEAD":
		return magenta
	case "OPTIONS":
		return white
	default:
		return reset
	}
}

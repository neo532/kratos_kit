package log

/*
 * @abstract 日志中间件
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/neo532/gofr/ghttp"

	"github.com/go-kratos/kratos/v2/errors"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/metadata"
	kmdw "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	httptr "github.com/go-kratos/kratos/v2/transport/http"

	"github.com/neo532/gokit/log"
	mdwServer "github.com/neo532/gokit/middleware/server"
)

var (
	AllowLogHeader = map[string]struct{}{
		"Content-Type": {},
		"User-Agent":   {},
	}
)

// Server is an server logging middleware.
func Server(logger klog.Logger) kmdw.Middleware {
	return func(handler kmdw.Handler) kmdw.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			ghttp.TagName = "json"

			var msg strings.Builder
			var path string
			startTime := time.Now()

			if info, ok := transport.FromServerContext(ctx); ok {

				// curl domain
				msg.WriteString("[curl ")
				msg.WriteString(info.Endpoint())

				// path
				path = info.Operation()
				if v, ok := info.(*httptr.Transport); ok {
					path = v.PathTemplate()
				}

				// method
				method := "POST"
				if m := info.RequestHeader().Get("Method"); m != "" {
					method = m
				}

				// curl path & param
				msg.WriteString(path)
				switch method {
				case "POST":

					var b []byte
					b, _ = json.Marshal(req)
					if len(b) != 0 && string(b) != "{}" {
						msg.WriteString(" -d '")
						msg.WriteString(string(b))
						msg.WriteString("'")
					}
				case "GET":

					s, _ := ghttp.Struct2QueryArgs(req)
					if len(s) > 0 {
						msg.WriteString("?" + s)
					}
				}

				msg.WriteString(" -X" + method)

				// curl header
				var header strings.Builder
				for _, h := range info.RequestHeader().Keys() {
					if _, ok := AllowLogHeader[h]; ok {
						if v := info.RequestHeader().Get(h); v != "" {
							header.WriteString(" -H '")
							header.WriteString(h)
							header.WriteString(":")
							header.WriteString(v)
							header.WriteString("'")
						}
					}
				}
				msg.WriteString(header.String())

				// curl end
				msg.WriteString("]")
			}

			// request reply
			msg.WriteString("[")
			reply, err = handler(ctx, req)
			var rstStr string
			if rst, err := json.Marshal(reply); err == nil {
				rstStr = string(rst)
				if mdwServer.IsProd(ctx) && len(rstStr) > log.MaxMsgLength {
					rstStr = rstStr[:log.MaxMsgLength]
				}
			}
			msg.WriteString(rstStr)
			msg.WriteString("]")

			// err
			level, stack, code, reason := extractError(err)

			md, ok := metadata.FromServerContext(ctx)
			if ok {
				if se := errors.FromError(err); se != nil {
					fmt.Println(runtime.Caller(0))
					mdt := make(map[string]string, len(md))
					md.Range(func(k string, v []string) bool {
						fmt.Println(fmt.Sprintf("v:\t%+v", v))
						mdt[k] = v[0]
						return true
					})
					err = se.WithMetadata(mdt)
				}
			}

			// if se := errors.FromError(err); se != nil {
			// 	fmt.Println(runtime.Caller(0))
			// 	err = se.WithMetadata(map[string]string{"aaa": "aaaaaaaaaaaa"})
			// }

			// log
			_ = klog.WithContext(ctx, logger).Log(level,
				"operation", path,
				"x_duration", time.Since(startTime).Seconds(),
				"x_curl", msg.String(),
				"stack", stack,
				"reason", reason,
				"code", code,
			)
			return
		}
	}
}

// extractError returns the string of the error
func extractError(err error) (level klog.Level, stack string, code int32, reason string) {
	if se := errors.FromError(err); se != nil {
		code = se.Code
		reason = se.Reason
	}
	if err != nil {
		level = klog.LevelError
		stack = fmt.Sprintf("%+v", err)
		return
	}
	level = klog.LevelInfo
	return
}

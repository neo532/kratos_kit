package tracing

/*
 * @abstract group信息传递中间件,用于分组调用
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"

	"github.com/neo532/gokit/middleware"

	"github.com/go-kratos/kratos/v2/log"
)

var (
	group string
)

func SetGroupForServer(ctx context.Context, gID string) context.Context {
	if gID == "" {
		gID = group
	}
	// 减少一次context赋值
	if id, ok := ctx.Value(middleware.Group).(string); ok && id == gID {
		return ctx
	}
	return context.WithValue(ctx, middleware.Group, gID)
}

func SetGroupForTracing(gID string) {
	group = gID
}

func GetGroupByCtx(ctx context.Context) string {
	if x, ok := ctx.Value(middleware.Group).(string); ok {
		return x
	}
	return group
}

// GetGroupForLog returns a Group valuer.
func GetGroupForLog() log.Valuer {
	return func(ctx context.Context) interface{} {
		if ctx == nil {
			ctx = context.Background()
		}
		return GetGroupByCtx(ctx)
	}
}

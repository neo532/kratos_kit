package tracing

/*
 * @abstract traceID信息传递中间件,用于服务跟踪标记
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"

	"github.com/neo532/gokit/middleware"

	"github.com/go-kratos/kratos/v2/log"
)

var name string

func SetNameForServer(ctx context.Context) context.Context {
	return context.WithValue(ctx, middleware.Name, name)
}

func SetNameForTracing(n string) {
	name = n
}

func GetFromByCtx(ctx context.Context) string {
	if x, ok := ctx.Value(middleware.From).(string); ok {
		return x
	}
	return ""
}

func GetEntryByCtx(ctx context.Context) string {
	if x, ok := ctx.Value(middleware.Entry).(string); ok {
		return x
	}
	return ""
}

func GetNameByCtx(ctx context.Context) string {
	if x, ok := ctx.Value(middleware.Name).(string); ok {
		return x
	}
	return ""
}

func SetFromForServer(ctx context.Context, from string) (c context.Context) {
	return context.WithValue(ctx, middleware.From, from)
}

// func SetTraceIDForClient(ori context.Context, dst context.Context) context.Context {
// 	var traceID string
// 	if tID, ok := ori.Value(KeyTraceID).(string); ok {
// 		traceID = tID
// 	}
// 	if traceID == "" {
// 		traceID = generateTraceID()
// 	}
// 	return context.WithValue(dst, KeyTraceID, traceID)
// }

// ========== for log ==========
// GetFromForLog returns a traceid valuer.
func GetFromForLog() log.Valuer {
	return func(ctx context.Context) interface{} {
		if ctx == nil {
			ctx = context.Background()
		}
		return GetFromByCtx(ctx)
	}
}

func GetEntryForLog() log.Valuer {
	return func(ctx context.Context) interface{} {
		if ctx == nil {
			ctx = context.Background()
		}
		return GetEntryByCtx(ctx)
	}
}

// ========== /for log ==========

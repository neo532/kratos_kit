package tracing

/*
 * @abstract traceID信息传递中间件,用于服务跟踪标记
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"
	"strings"

	"github.com/neo532/kratos_kit/middleware"

	"github.com/go-kratos/kratos/v2/log"
	uuid "github.com/satori/go.uuid"
)

func GetTraceIDByCtx(ctx context.Context) string {
	if x, ok := ctx.Value(middleware.TraceID).(string); ok {
		return x
	}
	return ""
}

func generateTraceID() string {
	return uuid.NewV4().String()
}

func SetTraceIDForServer(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		traceID = generateTraceID()
	}
	if strings.HasPrefix(traceID, "pts_") {
		ctx = context.WithValue(ctx, middleware.Benchmark, middleware.BenchmarkYes)
	}
	return context.WithValue(ctx, middleware.TraceID, traceID)
}

func IsBenchmark(ctx context.Context) (b bool) {
	if x, ok := ctx.Value(middleware.Benchmark).(string); ok {
		if x == middleware.BenchmarkYes {
			return true
		}
	}
	return
}

func UpdateNameByBenchmark(ctx context.Context, name string) string {
	if IsBenchmark(ctx) {
		return name + "_shadow"
	}
	return name
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
// TraceID returns a traceid valuer.
func GetTraceIDForLog() log.Valuer {
	return func(ctx context.Context) interface{} {
		if ctx == nil {
			ctx = context.Background()
		}
		return GetTraceIDByCtx(ctx)
	}
}

// ========== /for log ==========

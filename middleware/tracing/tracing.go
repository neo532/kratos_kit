package tracing

/*
 * @abstract trace信息传递中间件
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"

	"github.com/neo532/gokit/middleware"

	kmdw "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

func SetTraceKey(traceKey string) kmdw.Middleware {
	return func(handler kmdw.Handler) kmdw.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			return handler(ctx, req)
		}
	}
}

func Server() kmdw.Middleware {
	return func(handler kmdw.Handler) kmdw.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromServerContext(ctx); ok {

				// traceID
				traceID := tr.RequestHeader().Get(middleware.TraceID)
				ctx = SetTraceIDForServer(ctx, traceID)
				tr.RequestHeader().Set(middleware.TraceID, GetTraceIDByCtx(ctx))

				// rpcID
				rpcID := tr.RequestHeader().Get(middleware.RPCID)
				ctx = SetRpcIDForServer(ctx, rpcID)

				// group
				ctx = SetGroupForServer(ctx, "")

				// from
				from := tr.RequestHeader().Get(middleware.From)
				ctx = SetFromForServer(ctx, from)
			}
			return handler(ctx, req)
		}
	}
}

func Script(ctx context.Context) context.Context {
	// traceID
	traceID := GetTraceIDByCtx(ctx)
	ctx = SetTraceIDForServer(ctx, traceID)

	// rpcID
	rpcID := GetRpcIDByCtx(ctx)
	ctx = SetRpcIDForServer(ctx, rpcID)

	// group
	ctx = SetGroupForServer(ctx, "")

	// from
	from := GetFromByCtx(ctx)
	ctx = SetFromForServer(ctx, from)

	return ctx
}

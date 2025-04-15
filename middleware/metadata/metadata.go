package metadata

/*
 * @abstract metadata信息传递中间件
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"

	"github.com/neo532/kratos_kit/middleware"
	"github.com/neo532/kratos_kit/middleware/tracing"
)

// GetTraceIDByCtx returns traceID and value for metadata of client with context.
func GetTraceIDByCtx(ctx context.Context) (key, value string) {
	return middleware.TraceID, tracing.GetTraceIDByCtx(ctx)
}

// GetRpcIDByCtx returns rpcID and value for metadata of client with context.
func GetRpcIDByCtx(ctx context.Context) (key, value string) {
	return middleware.RPCID, tracing.GetRpcIDByCtx(ctx)
}

// GetGroupIDByCtx returns groupID and value for metadata of client with context.
func GetGroupIDByCtx(ctx context.Context) (key, value string) {
	return middleware.GroupID, tracing.GetGroupIDByCtx(ctx)
}

// GetNameByCtx returns name and value for metadata of client with context.
func GetNameByCtx(ctx context.Context) (key, value string) {
	return middleware.From, tracing.GetNameByCtx(ctx)
}

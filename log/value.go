package log

import (
	"context"
	"strconv"
	"time"

	"github.com/neo532/kratos_kit/middleware/tracing"

	klog "github.com/go-kratos/kratos/v2/log"
)

// Timestamp returns a timestamp Valuer with a custom time format.
func Timestamp() klog.Valuer {
	return func(context.Context) interface{} {
		return strconv.FormatInt(time.Now().Unix(), 10)
	}
}

// AddGlobalVariable returns logger with global variable.
func AddGlobalVariable(logger klog.Logger) klog.Logger {
	return klog.With(logger,
		"x_timestamp", Timestamp(),
		"x_traceid", tracing.GetTraceIDForLog(),
		"x_rpcid", tracing.GetRpcIDForLog(),
		"x_group", tracing.GetGroupForLog(),
		"x_from", tracing.GetFromForLog(),
		//"x_entry", tracing.GetEntryForLog(),
	)
}

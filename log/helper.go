package log

/*
 * @abstract 对接业务的日志具体操作
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"
	"fmt"
	"os"
	"time"

	klog "github.com/go-kratos/kratos/v2/log"
)

var (
	DefaultMessageKey = "x_msg"

	DefaultPathKey       = "path"
	DefaultDurationKey   = "x_duration"
	DefaultHTTPStatusKey = "x_http_status"
)

// Option is Helper option.
type Option func(*Helper)

// Helper is a logger helper.
type Helper struct {
	logger klog.Logger

	msgKey     string
	path       string
	cost       time.Duration
	httpstatus int
}

func WithMessageKey(k string) Option {
	return func(opts *Helper) {
		opts.msgKey = k
	}
}

// NewHelper new a logger helper.
func NewHelper(logger klog.Logger, opts ...Option) *Helper {
	options := &Helper{
		msgKey: DefaultMessageKey, // default message key
		logger: logger,
	}
	for _, o := range opts {
		o(options)
	}
	return options
}

// WithContext returns a shallow copy of h with its context changed
// to ctx. The provided ctx must be non-nil.
func (h *Helper) WithContext(ctx context.Context) *Helper {
	return &Helper{
		msgKey:     h.msgKey,
		path:       h.path,
		cost:       h.cost,
		httpstatus: h.httpstatus,
		logger:     klog.WithContext(ctx, h.logger),
	}
}

// WithPath returns a shallow copy of h with its module changed to module.
func (h *Helper) WithPath(p string) *Helper {
	return &Helper{
		msgKey:     h.msgKey,
		path:       p,
		cost:       h.cost,
		httpstatus: h.httpstatus,
		logger:     h.logger,
	}
}
func (h *Helper) WithCost(t time.Duration) *Helper {
	return &Helper{
		msgKey:     h.msgKey,
		path:       h.path,
		cost:       t,
		httpstatus: h.httpstatus,
		logger:     h.logger,
	}
}
func (h *Helper) WithHTTPStatus(code int) *Helper {
	return &Helper{
		msgKey:     h.msgKey,
		path:       h.path,
		cost:       h.cost,
		httpstatus: code,
		logger:     h.logger,
	}
}

// Log Print log by level and keyvals.
func (h *Helper) Log(level klog.Level, keyvals ...interface{}) {
	_ = h.logger.Log(level, keyvals...)
}

// Debug logs a message at debug level.
func (h *Helper) Debug(a ...interface{}) {
	h.Log(klog.LevelDebug, h.msgKey, fmt.Sprint(a...))
}

// Debugf logs a message at debug level.
func (h *Helper) Debugf(format string, a ...interface{}) {
	h.Log(klog.LevelDebug, h.msgKey, fmt.Sprintf(format, a...))
}

// Debugw logs a message at debug level.
func (h *Helper) Debugw(keyvals ...interface{}) {
	d := []interface{}{DefaultMessageKey}
	h.Log(klog.LevelDebug, append(d, keyvals)...)
}

// Info logs a message at info level.
func (h *Helper) Info(a ...interface{}) {
	if h.cost.Seconds() == 0 {
		h.Log(klog.LevelInfo, h.msgKey, fmt.Sprint(a...))
		return
	}
	h.Log(
		klog.LevelInfo,
		DefaultDurationKey, h.cost.Seconds(),
		DefaultPathKey, h.path,
		DefaultHTTPStatusKey, h.httpstatus,
		h.msgKey, fmt.Sprint(a...),
	)
}

// Infof logs a message at info level.
func (h *Helper) Infof(format string, a ...interface{}) {
	h.Log(klog.LevelInfo, h.msgKey, fmt.Sprintf(format, a...))
}

// Infow logs a message at info level.
func (h *Helper) Infow(keyvals ...interface{}) {
	d := []interface{}{DefaultMessageKey}
	h.Log(klog.LevelInfo, append(d, keyvals)...)
}

// Warn logs a message at warn level.
func (h *Helper) Warn(a ...interface{}) {
	h.Log(klog.LevelWarn, h.msgKey, fmt.Sprint(a...))
}

// Warnf logs a message at warnf level.
func (h *Helper) Warnf(format string, a ...interface{}) {
	h.Log(klog.LevelWarn, h.msgKey, fmt.Sprintf(format, a...))
}

// Warnw logs a message at warnf level.
func (h *Helper) Warnw(keyvals ...interface{}) {
	d := []interface{}{DefaultMessageKey}
	h.Log(klog.LevelWarn, append(d, keyvals)...)
}

// Error logs a message at error level.
func (h *Helper) Error(a ...interface{}) {
	if h.cost.Seconds() == 0 {
		h.Log(klog.LevelError, h.msgKey, fmt.Sprint(a...))
		return
	}
	h.Log(
		klog.LevelError,
		DefaultDurationKey, h.cost.Seconds(),
		DefaultPathKey, h.path,
		DefaultHTTPStatusKey, h.httpstatus,
		h.msgKey, fmt.Sprint(a...),
	)
}

// Errorf logs a message at error level.
func (h *Helper) Errorf(format string, a ...interface{}) {
	h.Log(klog.LevelError, h.msgKey, fmt.Sprintf(format, a...))
}

// Errorw logs a message at error level.
func (h *Helper) Errorw(keyvals ...interface{}) {
	d := []interface{}{DefaultMessageKey}
	h.Log(klog.LevelError, append(d, keyvals)...)
}

// Fatal logs a message at fatal level.
func (h *Helper) Fatal(a ...interface{}) {
	h.Log(klog.LevelFatal, h.msgKey, fmt.Sprint(a...))
	os.Exit(1)
}

// Fatalf logs a message at fatal level.
func (h *Helper) Fatalf(format string, a ...interface{}) {
	h.Log(klog.LevelFatal, h.msgKey, fmt.Sprintf(format, a...))
	os.Exit(1)
}

// Fatalw logs a message at fatal level.
func (h *Helper) Fatalw(keyvals ...interface{}) {
	d := []interface{}{DefaultMessageKey}
	h.Log(klog.LevelFatal, append(d, keyvals)...)
	os.Exit(1)
}

package zap

/*
 * @abstract zap的日志封装
 * @mail neo532@126.com
 * @date 2023-08-13
 */
import (
	"fmt"
	"os"

	//"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	klog "github.com/go-kratos/kratos/v2/log"
)

var _ klog.Logger = (*ZapLogger)(nil)

type ZapLogger struct {
	log  *zap.Logger
	Sync func() error
	err  error

	version    string
	ip         string
	department string
	name       string
	entry      string

	level zapcore.Level
	env   string

	syncerConf *lumberjack.Logger
}

type Option func(opt *ZapLogger)

// ========== option ==========
func WithLevel(l string) Option {
	return func(z *ZapLogger) {
		var err error
		if z.level, err = zapcore.ParseLevel(l); err != nil {
			z.err = err
		}
	}
}
func WithEnv(e string) Option {
	return func(z *ZapLogger) {
		z.env = e
	}
}
func WithVersion(s string) Option {
	return func(z *ZapLogger) {
		z.version = s
	}
}
func WithDepartment(s string) Option {
	return func(z *ZapLogger) {
		z.department = s
	}
}
func WithName(s string) Option {
	return func(z *ZapLogger) {
		z.name = s
	}
}
func WithEntry(s string) Option {
	return func(z *ZapLogger) {
		z.entry = s
	}
}
func WithIP(s string) Option {
	return func(z *ZapLogger) {
		z.ip = s
	}
}
func WithFilename(s string) Option {
	return func(z *ZapLogger) {
		z.syncerConf.Filename = s
	}
}

// 日志的最大大小（M）
func WithMaxSize(i int) Option {
	return func(z *ZapLogger) {
		z.syncerConf.MaxSize = i
	}
}

// 日志文件存储最大天数
func WithMaxAge(i int) Option {
	return func(z *ZapLogger) {
		z.syncerConf.MaxAge = i
	}
}

// 日志的最大保存数量
func WithMaxBackups(i int) Option {
	return func(z *ZapLogger) {
		z.syncerConf.MaxBackups = i
	}
}

// 是否执行压缩
func WithCompress(b bool) Option {
	return func(z *ZapLogger) {
		z.syncerConf.Compress = b
	}
}

// ========== /option ==========

func NewLogger(opts ...Option) (logger *ZapLogger) {
	logger = &ZapLogger{
		level:      zapcore.DebugLevel,
		env:        "dev",
		syncerConf: &lumberjack.Logger{},
	}
	for _, o := range opts {
		o(logger)
	}

	encoderConf := zapcore.EncoderConfig{
		LevelKey:    "x_level",
		TimeKey:     "x_date",
		CallerKey:   "x_line",
		FunctionKey: "x_func",

		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.NanosDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		LineEnding:     zapcore.DefaultLineEnding,
		//MessageKey: "messageKey",
		//NameKey:       "nameKey",
		//StacktraceKey: "stacktraceKey",
		//ConsoleSeparator: " | ",
	}

	var core zapcore.Core
	switch logger.env {

	case "dev":
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConf),
			zapcore.NewMultiWriteSyncer(
				zapcore.AddSync(os.Stdout),
				zapcore.AddSync(zapcore.AddSync(logger.syncerConf)),
			),
			logger.level,
		)

	default:
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConf),
			zapcore.NewMultiWriteSyncer(
				zapcore.AddSync(zapcore.AddSync(logger.syncerConf)),
			),
			logger.level,
		)
	}

	logger.log = zap.New(core,
		zap.WithCaller(true),
		zap.AddCallerSkip(4),
		zap.Fields(
			zap.String("x_name", logger.name),
			zap.String("x_version", logger.version),
			zap.String("x_server_ip", logger.ip),
			zap.String("x_department", logger.department),
			zap.String("x_entry", logger.entry),
		))
	logger.Sync = logger.log.Sync
	return
}

// Log 实现log接口
func (l *ZapLogger) Log(level klog.Level, keyvals ...interface{}) error {
	if len(keyvals) == 0 || len(keyvals)%2 != 0 {
		l.log.Warn(fmt.Sprint("Keyvalues must appear in pairs: ", keyvals))
		return nil
	}

	var data []zap.Field
	for i := 0; i < len(keyvals); i += 2 {
		data = append(data, zap.Any(fmt.Sprint(keyvals[i]), keyvals[i+1]))
	}

	switch level {
	case klog.LevelDebug:
		l.log.Debug("", data...)
	case klog.LevelInfo:
		l.log.Info("", data...)
	case klog.LevelWarn:
		l.log.Warn("", data...)
	case klog.LevelError:
		l.log.Error("", data...)
	case klog.LevelFatal:
		l.log.Fatal("", data...)
	}
	return nil
}

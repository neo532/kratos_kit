package redis

/*
 * @abstract Redis客户端
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"
	"strings"
	"sync"
	"time"

	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"

	"github.com/neo532/gokit/log"
)

// ========== Option ==========
type option struct {
	db          int
	goRedis     *redis.Options
	redisLogger *RedisLogger
	logger      *log.Helper
	ctx         context.Context
}
type Opt func(*option)

func WithMaxRetries(i int) Opt {
	return func(o *option) {
		o.goRedis.MaxRetries = i
	}
}
func WithReadTimeout(t time.Duration) Opt {
	return func(o *option) {
		o.goRedis.ReadTimeout = t
	}
}
func WithIdleTimeout(t time.Duration) Opt {
	return func(o *option) {
		o.goRedis.IdleTimeout = t
	}
}
func WithPoolSize(i int) Opt {
	return func(o *option) {
		o.goRedis.PoolSize = i
	}
}
func WithPassword(s string) Opt {
	return func(o *option) {
		o.goRedis.Password = s
	}
}
func WithDb(i int) Opt {
	return func(o *option) {
		o.goRedis.DB = i
	}
}
func WithSlowTime(t time.Duration) Opt {
	return func(o *option) {
		o.redisLogger.slowTime = t
	}
}
func WithLogger(l klog.Logger) Opt {
	return func(o *option) {
		o.logger = log.NewHelper(l)
		o.redisLogger.logger = o.logger
	}
}
func WithContext(c context.Context) Opt {
	return func(o *option) {
		o.ctx = c
	}
}

// ========== /Option ==========

var (
	instanceLock sync.Mutex
	redisMap     = make(map[string]*Redis, 2)
)

type Redis struct {
	Name    string
	Client  *redis.Client
	Cleanup func()
	Err     error
}

func New(name string, addr string, opts ...Opt) (rdb *Redis) {
	instanceLock.Lock()
	defer instanceLock.Unlock()

	var ok bool
	if rdb, ok = redisMap[name]; ok {
		return
	}

	opt := &option{
		goRedis: &redis.Options{
			Addr:        addr,
			PoolSize:    200,
			IdleTimeout: 240 * time.Second,
			ReadTimeout: 5 * time.Second,
			MaxRetries:  0,
		},
		logger: log.NewHelper(klog.DefaultLogger),
		ctx:    context.Background(),
		redisLogger: &RedisLogger{
			name:                 name,
			redisCtxBegintimeKey: "kit.database.redis_begintime",
			logger:               log.NewHelper(klog.DefaultLogger),
		},
	}
	for _, o := range opts {
		o(opt)
	}

	rdb = &Redis{
		Name:   name,
		Client: redis.NewClient(opt.goRedis),
	}
	rdb.Client.AddHook(opt.redisLogger)
	if rdb.Err = rdb.Client.Ping(opt.ctx).Err(); rdb.Err != nil {
		opt.logger.
			WithContext(opt.ctx).
			Errorf("New redis[%s] has err[err:%+v]!",
				name,
				rdb.Err,
			)
		return
	}

	rdb.Cleanup = func() {
		if rdb.Client == nil {
			opt.logger.
				WithContext(opt.ctx).
				Warnf("Close redis[%s] is nil!", name)
			return
		}
		if rdb.Err = rdb.Client.Close(); rdb.Err != nil {
			opt.logger.
				WithContext(opt.ctx).
				Warnf("Close redis[%s] has error: %+v",
					name,
					rdb.Err,
				)
			return
		}
	}

	redisMap[name] = rdb
	return
}

type RedisLogger struct {
	slowTime             time.Duration
	name                 string
	redisCtxBegintimeKey string
	logger               *log.Helper
}

func (h *RedisLogger) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, h.redisCtxBegintimeKey, time.Now()), nil
}

func (h *RedisLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) (err error) {

	// slow
	begin := ctx.Value(h.redisCtxBegintimeKey).(time.Time)
	cost := time.Since(begin)
	if cost > h.slowTime {
		h.logger.
			WithContext(ctx).
			Warnf("slowlog, name:%s, limit:%v, cost:%v, cmd:[%s]",
				h.name,
				h.slowTime,
				cost,
				cmd.String(),
			)
		return
	}

	// error
	if cmd.Err() != nil && cmd.Err() != redis.Nil {
		h.logger.
			WithContext(ctx).
			Errorf("err:[%+v], name:%s, limit:%v, cost:%v, cmd:[%s]",
				cmd.Err,
				h.name,
				h.slowTime,
				cost,
				cmd.String(),
			)
		return
	}

	// trace
	h.logger.
		WithContext(ctx).
		Infof("name:%s, limit:%v, cost:%v, cmd:[%s]",
			h.name,
			h.slowTime,
			cost,
			cmd.String(),
		)
	return
}

func (h *RedisLogger) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, h.redisCtxBegintimeKey, time.Now()), nil
}

func (h *RedisLogger) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) (err error) {
	var b strings.Builder
	for _, s := range cmds {
		b.WriteString(s.String() + ",")
	}

	// slow
	begin := ctx.Value(h.redisCtxBegintimeKey).(time.Time)
	cost := time.Since(begin)
	if cost > h.slowTime {
		h.logger.
			WithContext(ctx).
			Warnf("slowlog, name:%s, limit:%v, cost:%v, cmd:[%s]",
				h.name,
				h.slowTime,
				cost,
				b.String(),
			)
		return
	}

	// error
	for _, cmd := range cmds {
		if cmd.Err() != nil && cmd.Err() != redis.Nil {
			h.logger.
				WithContext(ctx).
				Errorf("err:[%+v], name:%s, limit:%v, cost:%v, cmd:[%s]",
					h.name,
					cmd.Err,
					h.slowTime,
					cost,
					cmd.String(),
				)
			return
		}
	}

	// trace
	h.logger.
		WithContext(ctx).
		Infof("name:%s, limit:%v, cost:%v, cmd:[%s]",
			h.name,
			h.slowTime,
			cost,
			b.String(),
		)
	return
}

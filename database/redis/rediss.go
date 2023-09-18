package redis

/*
 * @abstract Redis客户端
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"

	"github.com/go-redis/redis/v8"

	"github.com/neo532/gokit/middleware/server"
	"github.com/neo532/gokit/middleware/tracing"
)

type Rediss struct {
	def    *redis.Client
	shadow *redis.Client
	gray   *redis.Client

	cleanupFuncs []func()
	Err          error
}

func News(rdb *Redis) (rdbs *Rediss) {
	rdbs = &Rediss{}
	rdbs.def = rdbs.setClient(rdb)
	return
}

func (r *Rediss) SetShadow(rdb *Redis) *Rediss {
	r.shadow = r.setClient(rdb)
	return r
}

func (r *Rediss) SetGray(rdb *Redis) *Rediss {
	r.gray = r.setClient(rdb)
	return r
}

func (r *Rediss) Gray(c context.Context) (rdb *redis.Client) {
	if server.IsGray(c) && r.gray != nil {
		return r.gray
	}
	return r.Rdb(c)
}

func (r *Rediss) Rdb(c context.Context) (rdb *redis.Client) {
	if tracing.IsBenchmark(c) {
		return r.shadow
	}
	return r.def
}

func (r *Rediss) setClient(rdb *Redis) *redis.Client {
	r.cleanupFuncs = append(r.cleanupFuncs, rdb.Cleanup)
	if rdb.Err != nil {
		r.Err = rdb.Err
	}
	return rdb.Client
}

func (r *Rediss) Cleanup() func() {
	return func() {
		for _, fn := range r.cleanupFuncs {
			fn()
		}
	}
}

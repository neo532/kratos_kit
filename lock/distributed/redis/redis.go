package redis

/*
 * @abstract 用Redis实现分布式锁,此redis只能为单实例
 * @mail neo532@126.com
 * @date 2023-09-25
 */

import (
	"context"

	"github.com/neo532/kratos_kit/database/redis"
)

// ========== GoRedis ==========
type GoRedis struct {
	Rdb *redis.Rediss
}

func (l *GoRedis) Eval(c context.Context, cmd string, keys []string, args []interface{}) (rst interface{}, err error) {
	return l.Rdb.Rdb(c).Eval(c, cmd, keys, args...).Result()
}
func (l *GoRedis) Get(c context.Context, key string) (rst string, err error) {
	rst, err = l.Rdb.Rdb(c).Get(c, key).Result()
	return
}

// ========== /GoRedis ==========

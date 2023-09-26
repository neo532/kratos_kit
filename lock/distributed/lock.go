package lock

/*
 * @abstract 分布式锁的接口
 * @mail neo532@126.com
 * @date 2023-09-25
 */

import (
	"context"
	"time"
)

type DistributedLock interface {
	Lock(c context.Context, key string, expire, wait time.Duration) (code string, err error)
	UnLock(c context.Context, key string, code string) (err error)
}

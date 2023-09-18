package tracing

/*
 * @abstract rpcID信息传递中间件,用于服务调用标记
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/neo532/gokit/middleware"
)

// ========== RpcID==========
type RpcID struct {
	id   atomic.Value
	lock sync.Mutex
	key  string
}

// AddRpcIDSibling
// 在本服务客户端调用的时候
func (r *RpcID) AddSibling(ctx context.Context) context.Context {
	r.lock.Lock()
	defer r.lock.Unlock()

	id := r.Get()

	// empty string, eg: ""
	if id == "" {
		return r.Set(ctx, "1")
	}

	// don't have dot.eg: "1"
	index := strings.LastIndex(id, ".")
	if index == -1 {
		if ver, err := strconv.Atoi(id); err == nil {
			return r.Set(ctx, strconv.Itoa(ver+1))
		}
		return ctx
	}

	// normal, eg: "1.1.0"
	if ver, err := strconv.Atoi(id[index+1:]); err == nil {
		return r.Set(ctx, id[:index+1]+strconv.Itoa(ver+1))
	}

	return ctx
}

func (r *RpcID) Key(key string) *RpcID {
	r.key = key
	return r
}

func (r *RpcID) Set(ctx context.Context, id string) context.Context {
	var key = r.key
	if key == "" {
		key = middleware.RPCID
	}
	r.id.Store(id)
	return context.WithValue(ctx, key, r)
}

func (r *RpcID) Get() string {
	id := r.id.Load()
	if i, ok := id.(string); ok {
		return i
	}
	return ""
}

// AddRpcIDLayer 添加一层.eg:""=>0, 1.1=>1.1.0
// 无锁，在外部请求进入内部服务的时候，调一下。
func (r *RpcID) AddLayer(ctx context.Context, rpcID string) context.Context {
	return r.Set(ctx, strings.TrimPrefix(rpcID+".0", "."))
}

// ========== /RpcID==========

func GetRpcIDByCtx(ctx context.Context) string {
	if x, ok := ctx.Value(middleware.RPCID).(*RpcID); ok {
		return x.Get()
	}
	return ""
}

func SetRpcIDForClient(ctx context.Context) context.Context {
	if x, ok := ctx.Value(middleware.RPCID).(*RpcID); ok {
		return x.AddSibling(ctx)
	}

	return (&RpcID{}).AddSibling(ctx)
}

// 为本服务添加rpcID初始化
func SetRpcIDForServer(ctx context.Context, rpcID string) context.Context {
	// 减少一次context复制
	if x, ok := ctx.Value(middleware.RPCID).(*RpcID); ok {
		return x.AddLayer(ctx, rpcID)
	}

	return (&RpcID{}).AddLayer(ctx, rpcID)
}

// RpcID returns a rpcid valuer.
func GetRpcIDForLog() log.Valuer {
	return func(ctx context.Context) interface{} {
		if ctx == nil {
			ctx = context.Background()
		}
		return GetRpcIDByCtx(ctx)
	}
}

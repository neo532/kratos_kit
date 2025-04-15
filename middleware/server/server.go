package server

/*
 * @abstract 服务器相关中间件
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"

	kmdw "github.com/go-kratos/kratos/v2/middleware"

	"github.com/neo532/kratos_kit/middleware"
)

// Env is an server logging middleware.
func SetEnv(env string) kmdw.Middleware {
	return func(handler kmdw.Handler) kmdw.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {

			if _, ok := ctx.Value(middleware.Env).(string); !ok {
				ctx = context.WithValue(ctx, middleware.Env, env)
			}
			return handler(ctx, req)
		}
	}
}
func Env(ctx context.Context) (env string) {
	env, _ = ctx.Value(middleware.Env).(string)
	return
}
func IsGray(ctx context.Context) bool {
	return Env(ctx) == middleware.EnvGray
}
func IsProd(ctx context.Context) bool {
	return Env(ctx) == middleware.EnvProd
}
func IsDev(ctx context.Context) bool {
	return Env(ctx) == middleware.EnvDev
}

func SetEntry(entry string) kmdw.Middleware {
	return func(handler kmdw.Handler) kmdw.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {

			if _, ok := ctx.Value(middleware.Entry).(string); !ok {
				ctx = context.WithValue(ctx, middleware.Entry, entry)
			}
			return handler(ctx, req)
		}
	}
}

func Entry(ctx context.Context) (entry string) {
	entry, _ = ctx.Value(middleware.Entry).(string)
	return
}

// ========= tmp =========
func SetEntryForCtx(ctx context.Context, entry string) context.Context {
	return context.WithValue(ctx, middleware.Entry, entry)
}
func SetEnvForCtx(ctx context.Context, env string) context.Context {
	return context.WithValue(ctx, middleware.Env, env)
}

// ========= /tmp =========

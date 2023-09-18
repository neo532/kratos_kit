package kafka

/*
 * @abstract kafka的具体参数定义
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/neo532/gokit/log"
)

type logger struct {
	log *log.Helper
	ctx context.Context
}

func NewLogger(l klog.Logger, module string) *logger {
	lo := &logger{
		log: log.NewHelper(l),
		ctx: context.Background(),
	}
	return lo
}
func (l *logger) WithContext(c context.Context) {
	l.ctx = c
}
func (l *logger) Print(v ...interface{}) {
	l.log.WithContext(l.ctx).Info(v...)
}
func (l *logger) Printf(format string, v ...interface{}) {
	l.log.WithContext(l.ctx).Infof(format, v...)
}
func (l *logger) Println(v ...interface{}) {
	l.log.WithContext(l.ctx).Info(v...)
}

package queue

/*
 * @abstract 消息队列消费者的接口定义
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"
	"strings"
)

type Consumer interface {
	Start(context.Context) error
	Stop(context.Context) error
	Name() string
}

// ========== Consumers ==========
type Consumers struct {
	csm map[string]Consumer
}

func NewConsumers(cs ...Consumer) Consumer {
	groups := &Consumers{
		csm: make(map[string]Consumer, len(cs)),
	}
	for _, o := range cs {
		groups.csm[o.Name()] = o
	}
	return groups
}
func (cs *Consumers) Start(ctx context.Context) (err error) {
	for _, o := range cs.csm {
		if e := o.Start(ctx); e != nil {
			err = e
		}
	}
	return
}
func (cs *Consumers) Stop(ctx context.Context) (err error) {
	for _, o := range cs.csm {
		if e := o.Stop(ctx); e != nil {
			err = e
		}
	}
	return
}
func (cs *Consumers) Name() (name string) {
	for _, o := range cs.csm {
		name += "," + o.Name()
	}
	return strings.TrimPrefix(name, ",")
}

// ========== /Consumers ==========

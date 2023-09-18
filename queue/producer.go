package queue

/*
 * @abstract 消息队列生产者的接口定义
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"

	"github.com/neo532/gokit/middleware/server"
	"github.com/neo532/gokit/middleware/tracing"
)

type Producer interface {
	Err() error
	CleanUp() func()
	Send(c context.Context, message []byte, hashKey ...string) error
}

// ========== Producer ==========
type Producers struct {
	def    Producer
	shadow Producer
	gray   Producer

	cleanUpFuncs []func()
	Err          error
}

func NewProducers(p Producer) (pdc *Producers) {
	pdc = &Producers{}
	pdc.def = pdc.setClient(p)
	return
}

func (p *Producers) SetShadow(pdc Producer) *Producers {
	p.shadow = p.setClient(pdc)
	return p
}

func (p *Producers) SetGray(pdc Producer) *Producers {
	p.gray = p.setClient(pdc)
	return p
}

func (p *Producers) Gray(c context.Context) Producer {
	if server.IsGray(c) && p.gray != nil {
		return p.gray
	}
	return p.Producer(c)
}

func (p *Producers) Producer(c context.Context) Producer {
	if tracing.IsBenchmark(c) {
		return p.shadow
	}
	return p.def
}

func (p *Producers) setClient(pdc Producer) Producer {
	p.cleanUpFuncs = append(p.cleanUpFuncs, pdc.CleanUp())
	if pdc.Err() != nil {
		p.Err = pdc.Err()
	}
	return pdc
}

func (p *Producers) CleanUp() func() {
	return func() {
		for _, fn := range p.cleanUpFuncs {
			fn()
		}
	}
}

// ========== /Producers ==========

package producer

/*
 * @abstract 消息队列kafka的消息生产端的实现
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"
	"sync"
	"time"

	"github.com/IBM/sarama"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/neo532/kratos_kit/log"
	"github.com/neo532/kratos_kit/middleware"
	"github.com/neo532/kratos_kit/middleware/tracing"
	"github.com/neo532/kratos_kit/queue/kafka"
)

var (
	instanceLock sync.Mutex
	producerMap  = make(map[string]*Producer, 2)
)

// ========== Option ==========
type Opt func(*Producer)

func WithVersion(ver sarama.KafkaVersion) Opt {
	return func(o *Producer) {
		o.conf.Version = ver
	}
}

func WithLogger(l klog.Logger, ctx ...context.Context) Opt {
	return func(o *Producer) {
		o.logger = log.NewHelper(l)
		sl := kafka.NewLogger(l, "sarama")
		if len(ctx) > 0 {
			sl.WithContext(ctx[0])
		}
		sarama.DebugLogger = sl
	}
}

func WithSync(b bool) Opt {
	return func(o *Producer) {
		o.isSync = b
	}
}

func WithRequiredAcks(r sarama.RequiredAcks) Opt {
	return func(o *Producer) {
		o.conf.Producer.RequiredAcks = r
	}
}

func WithReturnSucesses(b bool) Opt {
	return func(o *Producer) {
		o.conf.Producer.Return.Successes = b
	}
}

// sarama.NewHashPartitioner
func WithPartitioner(fn sarama.PartitionerConstructor) Opt {
	return func(o *Producer) {
		o.conf.Producer.Partitioner = fn
	}
}

func WithTopic(topic string) Opt {
	return func(o *Producer) {
		o.topic = topic
	}
}

func WithContext(c context.Context) Opt {
	return func(o *Producer) {
		o.bootstrapContext = c
	}
}
func WithIdempotent(b bool) Opt {
	return func(o *Producer) {
		o.conf.Producer.Idempotent = b
	}
}
func WithNetMaxOpenRequest(i int) Opt {
	return func(o *Producer) {
		o.conf.Net.MaxOpenRequests = i
	}
}

// ========== /Option ==========

type Producer struct {
	name   string
	conf   *sarama.Config
	addrs  []string
	logger *log.Helper

	cleanUp          func()
	err              error
	bootstrapContext context.Context

	syncProducer  sarama.SyncProducer
	asyncProducer sarama.AsyncProducer

	isSync bool
	topic  string
}

func New(name string, addrs []string, opts ...Opt) (pdc *Producer) {
	// one instance
	instanceLock.Lock()
	defer instanceLock.Unlock()
	var ok bool
	if pdc, ok = producerMap[name]; ok {
		return
	}

	// init
	pdc = &Producer{
		name:             name,
		conf:             sarama.NewConfig(),
		addrs:            addrs,
		logger:           log.NewHelper(klog.DefaultLogger),
		bootstrapContext: context.Background(),
	}
	pdc.conf.Version = sarama.V0_11_0_2
	for _, o := range opts {
		o(pdc)
	}

	// validate
	if pdc.err = pdc.conf.Validate(); pdc.err != nil {
		pdc.logger.
			WithContext(pdc.bootstrapContext).
			Errorf("conf.Validate[%s] Has error.[conf:%+v][err:%+v]", pdc.name, pdc.conf, pdc.err)
		return
	}

	switch pdc.isSync {
	case true:
		pdc.syncProducer, pdc.err = sarama.NewSyncProducer(pdc.addrs, pdc.conf)
		pdc.cleanUp = func() {
			if pdc.syncProducer != nil {
				pdc.err = pdc.syncProducer.Close()
			}
		}
	case false:
		pdc.asyncProducer, pdc.err = sarama.NewAsyncProducer(pdc.addrs, pdc.conf)
		pdc.cleanUp = func() {
			if pdc.asyncProducer != nil {
				pdc.err = pdc.asyncProducer.Close()
			}
		}
		go func(producer sarama.AsyncProducer) {
			for {
				select {
				case e := <-producer.Errors():
					if e != nil {
						pdc.logger.
							WithContext(pdc.bootstrapContext).
							Errorf(
								"Async producer has error![err:%s][topic:%s][offset:%d][partition:%d][key:%v][value:%s]",
								e.Error(),
								e.Msg.Topic,
								e.Msg.Offset,
								e.Msg.Partition,
								e.Msg.Key,
								e.Msg.Value,
							)
					}
				case <-producer.Successes():
				case <-pdc.bootstrapContext.Done():
					return
				}
			}
		}(pdc.asyncProducer)
	}
	if pdc.err != nil {
		pdc.logger.
			WithContext(pdc.bootstrapContext).
			Errorf("sarama.NewProducer[name:%s][isSync:%b] has error![err:%+v]", pdc.name, pdc.isSync, pdc.err)
		return
	}
	pdc.logger.
		WithContext(pdc.bootstrapContext).
		Infof("Producer[name:%s] Running!", pdc.name)
	return
}

func (pdc *Producer) Err() error {
	return pdc.err
}

func (pdc *Producer) Send(ctx context.Context, message []byte, hashKey ...string) (err error) {
	var hKey string
	pm := &sarama.ProducerMessage{
		Topic:     pdc.topic,
		Value:     sarama.StringEncoder(string(message)),
		Timestamp: time.Now(),
		Headers: []sarama.RecordHeader{
			{[]byte(middleware.TraceID), []byte(tracing.GetTraceIDByCtx(ctx))},
			{[]byte(middleware.RPCID), []byte(tracing.GetRpcIDByCtx(ctx))},
			{[]byte(middleware.Group), []byte(tracing.GetGroupByCtx(ctx))},
		}, // at lease kafka v0.11+
	}
	if len(hashKey) > 0 {
		hKey = hashKey[0]
		pm.Key = sarama.StringEncoder(hKey)
	}

	switch pdc.isSync {
	case true:
		_, _, err = pdc.syncProducer.SendMessage(pm)
	case false:
		pdc.asyncProducer.Input() <- pm
	}

	if err != nil {
		pdc.logger.
			WithContext(ctx).
			Errorf("Producer[%s]'s sending Has err![err:%s][hkey:%s][msg:%s]",
				pdc.name,
				err,
				hKey,
				message,
			)
		return
	}

	pdc.logger.
		WithContext(ctx).
		Infof("Producer[%s] have been delivered![hkey:%s][msg:%s]",
			pdc.name,
			hKey,
			message,
		)
	return
}

func (pdc *Producer) CleanUp() func() {
	return pdc.cleanUp
}

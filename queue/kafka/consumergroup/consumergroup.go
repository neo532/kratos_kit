package consumergroup

/*
 * @abstract 消息队列kafka的按组消费的客户端实现
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/IBM/sarama"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/neo532/gokit/log"
	"github.com/neo532/gokit/middleware"
	"github.com/neo532/gokit/middleware/tracing"
	"github.com/neo532/gokit/queue/kafka"
)

var (
	instanceLock sync.Mutex
	minGoCount   = 3
)

// ========== Option ==========
type Opt func(o *ConsumerGroup)

func WithAutoCommit(b bool) Opt {
	return func(o *ConsumerGroup) {
		o.conf.Consumer.Offsets.AutoCommit.Enable = b
	}
}
func WithBalanceStrategy(strategy sarama.BalanceStrategy) Opt {
	return func(o *ConsumerGroup) {
		o.conf.Consumer.Group.Rebalance.Strategy = strategy
	}
}
func WithVersion(ver sarama.KafkaVersion) Opt {
	return func(o *ConsumerGroup) {
		o.conf.Version = ver
	}
}

func WithLogger(l klog.Logger, ctx ...context.Context) Opt {
	return func(o *ConsumerGroup) {
		sl := kafka.NewLogger(l, "sarama")
		if len(ctx) > 0 {
			sl.WithContext(ctx[0])
		}
		sarama.DebugLogger = sl

		o.logger = log.NewHelper(l)
		o.handler.logger = o.logger
	}
}
func WithSlowLog(t time.Duration) Opt {
	return func(o *ConsumerGroup) {
		o.handler.slowTime = t
	}
}
func WithGoCount(count int) Opt {
	return func(o *ConsumerGroup) {
		o.goCount = count
	}
}
func WithEnv(env string) Opt {
	return func(o *ConsumerGroup) {
		o.handler.env = env
	}
}

//	func WithDelimiter(s string) Opt {
//		return func(o *ConsumerGroup) {
//			o.delimiter = s
//		}
//	}
func WithTopics(s ...string) Opt {
	return func(o *ConsumerGroup) {
		o.topics = s
	}
}
func WithCallback(fn func(ctx context.Context, message []byte) (err error)) Opt {
	return func(o *ConsumerGroup) {
		o.handler.callback = fn
	}
}
func WithContext(c context.Context) Opt {
	return func(o *ConsumerGroup) {
		o.bootstrapContext = c
	}
}

// ========== /Option ==========

// ========== ConsumerGroup ==========
type ConsumerGroup struct {
	name   string
	conf   *sarama.Config
	addrs  []string
	group  string
	logger *log.Helper
	err    error

	goCount          int
	bootstrapContext context.Context

	topics  []string
	handler *groupHandler

	//delimiter string

	consumer sarama.ConsumerGroup
}

func NewGroup(name string, addrs []string, group string, opts ...Opt) (csm *ConsumerGroup) {

	// init parameter
	logger := log.NewHelper(klog.DefaultLogger)
	csm = &ConsumerGroup{
		name:    name,
		conf:    sarama.NewConfig(),
		addrs:   addrs,
		group:   group,
		logger:  logger,
		goCount: runtime.NumCPU() / 2,
		handler: &groupHandler{
			name:     name,
			slowTime: 3 * time.Second,
			logger:   logger,
		},
		bootstrapContext: context.Background(),
	}
	if csm.goCount < minGoCount {
		csm.goCount = minGoCount
	}
	csm.conf.Version = sarama.V0_11_0_2
	for _, o := range opts {
		o(csm)
	}
	csm.handler.autoCommit = csm.conf.Consumer.Offsets.AutoCommit.Enable
	csm.conf.Consumer.MaxWaitTime = time.Second

	// check
	if csm.err = csm.conf.Validate(); csm.err != nil {
		csm.logger.
			WithContext(csm.bootstrapContext).
			Errorf("Validate has error.[conf:%+v][err:%+v]", csm.conf, csm.err)
		return
	}

	// initilize
	if csm.consumer, csm.err = sarama.NewConsumerGroup(csm.addrs, csm.group, csm.conf); csm.err != nil {
		csm.logger.
			WithContext(csm.bootstrapContext).
			Errorf("NewGroup has error![err:%+v]", csm.err)
		return
	}

	return
}

func (csm *ConsumerGroup) Name() (name string) {
	return csm.name
}

func (csm *ConsumerGroup) Stop(ctx context.Context) (err error) {
	if csm.consumer != nil {
		err = csm.consumer.Close()
	}
	return
}

func (csm *ConsumerGroup) Start(ctx context.Context) (err error) {
	for i := 0; i < csm.goCount; i++ {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					csm.logger.
						WithContext(ctx).
						Errorf("Start has panic![err:%+v][stack:%s]", err, string(debug.Stack()))
				}
			}()

			for {
				select {
				case <-ctx.Done():
					csm.logger.
						WithContext(ctx).
						Infof("topic %v consumer have canceled!", csm.topics)
					return
				default:
					csm.logger.
						WithContext(ctx).
						Infof("[name:%s],[topics:%s],[addrs:%s],[group:%s] consumer is starting!",
							csm.name, strings.Join(csm.topics, ","), strings.Join(csm.addrs, ","), csm.group,
						)

					// 此方法会一直阻塞，一直到消费者服务停掉
					if err := csm.consumer.Consume(ctx, csm.topics, csm.handler); err != nil {
						csm.logger.
							WithContext(ctx).
							Errorf("Consume has error![err:%v]", err)
						return
					}
				}
			}
		}()
	}
	return
}

// ========== /ConsumerGroup ==========

// ========== instance of kafka handle ==========
// Consumer represents a Sarama consumer group consumer
type groupHandler struct {
	env        string
	name       string
	autoCommit bool
	callback   func(ctx context.Context, message []byte) (err error)
	slowTime   time.Duration
	logger     *log.Helper
	ctx        context.Context
	msg        []byte
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *groupHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *groupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (h *groupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) (err error) {
	defer func() {
		if err := recover(); err != nil {
			h.logger.
				WithContext(h.ctx).
				Errorf("Handler has panic![err:%+v][msg:%s][trace:%s][stack:%s]", err, h.msg, h.getTrace(h.ctx), string(debug.Stack()))
		}
		time.Sleep(time.Second)
		// TODO 这里可以搞个错误队列，不过目前只做了延迟1秒钟继续处理，会一直堵在这，直到消费成功为止
		// TODO 这里可以先做个失败重试次数，重试多次后，就提交防止一直阻塞。
	}()

	for msg := range claim.Messages() {

		h.ctx = h.setTrace(session.Context(), msg.Headers)
		h.msg = msg.Value

		begin := time.Now()
		err := h.callback(h.ctx, msg.Value)
		cost := time.Since(begin)
		// mark ok
		session.MarkMessage(msg, "")
		// biz error
		if err != nil {
			h.logger.
				WithContext(h.ctx).
				Errorf("callback has error![name:%s][err:%v][trace:%s][msg:%s]", h.name, err, h.getTrace(h.ctx), msg.Value)
			continue
		}

		if !h.autoCommit {
			session.Commit()
		}

		// slow
		if cost > h.slowTime {
			h.logger.
				WithContext(h.ctx).
				Warnf("slowlog [name:%s], [limit:%v], [cost:%v], [trace:%s], [msg:%s]", h.name, h.slowTime, cost, h.getTrace(h.ctx), msg.Value)
		}

		if h.env == middleware.EnvProd && utf8.RuneCount(msg.Value) > log.MaxMsgLength {
			msg.Value = []byte(string([]rune(string(msg.Value))[:log.MaxMsgLength]) + "...")
		}
		h.logger.
			WithContext(h.ctx).
			Infof("[name:%s], [partition:%d], [offset:%d], [cost:%v], [trace:%s], [msg:%s]",
				h.name,
				msg.Partition,
				msg.Offset,
				cost,
				h.getTrace(h.ctx),
				msg.Value,
			)
	}
	return nil
}

func (h *groupHandler) setTrace(ctx context.Context, header []*sarama.RecordHeader) context.Context {
	for _, h := range header {
		switch string(h.Key) {
		case middleware.TraceID:
			ctx = tracing.SetTraceIDForServer(ctx, string(h.Value))
		case middleware.RPCID:
			ctx = tracing.SetRpcIDForServer(ctx, string(h.Value))
		case middleware.Group:
			ctx = tracing.SetGroupForServer(ctx, string(h.Value))
		case middleware.From:
			ctx = tracing.SetFromForServer(ctx, string(h.Value))
		}
	}
	return ctx
}

func (h *groupHandler) getTrace(ctx context.Context) (trace string) {
	return tracing.GetGroupByCtx(ctx) + ";" + tracing.GetRpcIDByCtx(ctx) + ";" + tracing.GetTraceIDByCtx(ctx)
}

// ========== /instance of kafka handle ==========

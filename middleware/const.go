package middleware

/*
 * @abstract 中间件key定义
 * @mail neo532@126.com
 * @date 2023-08-13
 */

// context key
const (
	Env       = "env"       // 环境
	Entry     = "entry"     // 入口:api|script|consumer
	From      = "from"      // 从哪个服务请求来的
	Name      = "name"      // 当前服务名
	Benchmark = "benchmark" // 是否是压测
	Timestamp = "timestamp" // 用于返回值

	TraceID     = "traceId"     // 链路跟踪ID
	RPCID       = "rpcId"       //use for client header to request other service.
	RPCIDServer = "rpcIdServer" //use for logging.
	Group       = "group"       //use for dev.
)

// context value
const (
	EnvDev  = "dev"
	EnvTest = "test"
	EnvProd = "prod"
	EnvGray = "gray"

	EntryApi      = "api"
	EntryAdmin    = "admin"
	EntryScript   = "script"
	EntryConsumer = "consumer"

	BenchmarkYes = "1"
)

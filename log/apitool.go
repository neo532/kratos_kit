package log

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"
	//"github.com/neo532/apitool/transport/http/xhttp"
)

func NewXHttpLogger(log klog.Logger) *XHttpLogger {
	return &XHttpLogger{
		log: NewHelper(log),
	}
}

type XHttpLogger struct {
	log *Helper
}

func (l *XHttpLogger) Info(c context.Context, msg string) {
	// if info := xhttp.GetCurlInfoByCtx(c); info != nil {
	// 	l.log.
	// 		WithContext(c).
	// 		WithCost(info.Cost).
	// 		WithPath(info.Url).
	// 		WithHTTPStatus(info.HTTPStatus).
	// 		Info(msg)
	// 	return
	// }
	l.log.WithContext(c).Info(msg)
	return
}

func (l *XHttpLogger) Error(c context.Context, msg string) {
	// if info := xhttp.GetCurlInfoByCtx(c); info != nil {
	// 	l.log.
	// 		WithContext(c).
	// 		WithCost(info.Cost).
	// 		WithPath(info.Url).
	// 		WithHTTPStatus(info.HTTPStatus).
	// 		Error(msg)
	// 	return
	// }
	l.log.WithContext(c).Error(msg)
	return
}

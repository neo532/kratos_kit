package http

/*
 * @abstract 传输协议http的响应体包装
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/neo532/gokit/middleware"

	proto "github.com/golang/protobuf/proto"
	anypb "google.golang.org/protobuf/types/known/anypb"

	"github.com/go-kratos/kratos/v2/errors"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// ContentType returns the content-type with base prefix.
func ContentType(subtype string) string {
	return strings.Join([]string{"application", subtype}, "/")
}

func IsCallback(url string) bool {
	return strings.HasSuffix(url, "/callback")
}

func IsOriginCallback(url string) bool {
	if strings.Index(url, "/callback/origin") != -1 {
		return true
	}
	return false
}

type OriginCallback struct {
	Data string `json:"data"`
}

func ResponseEncoder(w http.ResponseWriter, r *http.Request, d interface{}) (err error) {
	codec, _ := khttp.CodecForRequest(r, "Accept")
	w.Header().Set("Content-Type", ContentType(codec.Name()))

	traceId := r.Header.Get(middleware.TraceID)
	w.Header().Set(middleware.TraceID, traceId)
	var data []byte
	switch {
	case IsOriginCallback(r.RequestURI):
		if data, err = codec.Marshal(d); err != nil {
			return
		}
		var ori = new(OriginCallback)
		if err = codec.Unmarshal(data, &ori); err != nil {
			return
		}
		data = []byte(ori.Data)
	case IsCallback(r.RequestURI):
		if data, err = codec.Marshal(d); err != nil {
			return
		}
	default:
		reply := &Response{
			Code:         0,
			Message:      "ok",
			TraceId:      traceId,
			ResponseTime: uint32(time.Now().Unix()),
		}
		if v, ok := d.(proto.Message); ok {
			any, e := anypb.New(proto.MessageV2(v))
			if e == nil {
				reply.Data = any
			}
		}
		if data, err = codec.Marshal(reply); err != nil {
			return
		}
	}

	_, err = w.Write(data)
	return
}

func ErrorEncoder(w http.ResponseWriter, r *http.Request, err error) {
	codec, _ := khttp.CodecForRequest(r, "Accept")
	w.Header().Set("Content-Type", ContentType(codec.Name()))
	traceId := r.Header.Get(middleware.TraceID)
	w.Header().Set(middleware.TraceID, traceId)
	se := errors.FromError(err)
	code := int(se.Code)
	var body []byte

	switch {
	case IsCallback(r.RequestURI):
		response := strings.Replace(se.Message, "{code}", strconv.Itoa(code), 1)
		response = strings.Replace(response, "{reason}", se.Reason, 1)
		body = []byte(response)

	default:
		reply := &Response{
			Code:         int32(code),
			Message:      se.Message,
			Reason:       err.Error(),
			Data:         nil,
			TraceId:      traceId,
			ResponseTime: uint32(time.Now().Unix()),
		}
		body, err = codec.Marshal(reply)
	}

	if err != nil && code < 600 {
		w.WriteHeader(code)
	}
	_, _ = w.Write(body)
	return
}

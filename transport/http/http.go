package http

/*
 * @abstract 传输协议http的响应体包装
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"github.com/go-kratos/kratos/v2/encoding/json"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	json.MarshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
	}
}

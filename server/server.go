package server

/*
 * @abstract 一些关于服务的操作
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

// WrapEr,err可能为nil,er肯定不为nil
func WrapEr(err error, er error, message string) error {
	if err == nil {
		return errors.Wrap(er, message)
	}
	return errors.Wrap(err, fmt.Sprintf(" [%s] [%s]", er.Error(), message))
}

func WritePID(pid int, file string) (err error) {
	p := strconv.Itoa(pid)
	if file == "" {
		file = "./pid"
	}

	var f *os.File
	f, err = os.OpenFile(file, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	defer f.Close()
	if err != nil {
		return
	}
	var n int
	n, err = f.Write([]byte(p))
	if err != nil {
		return
	}
	if n < len(p) {
		err = io.ErrShortWrite
	}
	return
}

package cbc

import (
	"encoding/base64"

	"github.com/forgoer/openssl"
)

type CBC struct {
	padding string
	key     []byte
	iv      []byte
}

type opt func(o *CBC)

func WithPadding(padding string) opt {
	return func(o *CBC) {
		o.padding = padding
	}
}
func WithKey(key []byte) opt {
	return func(o *CBC) {
		o.key = key
	}
}
func WithIv(iv []byte) opt {
	return func(o *CBC) {
		o.iv = iv
	}
}

func New(opts ...opt) (os *CBC) {
	os = &CBC{
		padding: openssl.PKCS7_PADDING,
	}
	for _, fn := range opts {
		fn(os)
	}
	return os
}

func (o *CBC) Encrypt(origin []byte) (encrypt string, err error) {
	var en []byte
	en, err = openssl.AesCBCEncrypt(origin, o.key, o.iv, o.padding)
	encrypt = base64.StdEncoding.EncodeToString(en)
	return
}

func (o *CBC) Decrypt(encrypt string) (origin []byte, err error) {
	return openssl.AesCBCEncrypt([]byte(encrypt), o.key, o.iv, o.padding)
}

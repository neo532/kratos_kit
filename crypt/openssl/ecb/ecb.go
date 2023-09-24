package ecb

import (
	"encoding/base64"

	"github.com/forgoer/openssl"
)

type ECB struct {
	padding string
	key     []byte
}

type opt func(o *ECB)

func WithPadding(padding string) opt {
	return func(o *ECB) {
		o.padding = padding
	}
}
func WithKey(key []byte) opt {
	return func(o *ECB) {
		o.key = key
	}
}

func New(opts ...opt) (os *ECB) {
	os = &ECB{
		padding: openssl.PKCS7_PADDING,
	}
	for _, fn := range opts {
		fn(os)
	}
	return os
}

func (o *ECB) Encrypt(origin []byte) (encrypt string, err error) {
	var en []byte
	en, err = openssl.AesECBEncrypt(origin, o.key, o.padding)
	encrypt = base64.StdEncoding.EncodeToString(en)
	return
}

func (o *ECB) Decrypt(encrypt string) (origin []byte, err error) {
	var en []byte
	en, err = base64.StdEncoding.DecodeString(encrypt)
	return openssl.AesECBDecrypt(en, o.key, o.padding)
}

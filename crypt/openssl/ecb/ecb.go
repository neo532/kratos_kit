package ecb

import (
	"github.com/forgoer/openssl"

	"github.com/neo532/gokit/crypt"
	"github.com/neo532/gokit/crypt/encoding/std"
)

type ECB struct {
	padding string
	key     []byte
	coding  crypt.IEncoding
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
func WithEncoding(coding crypt.IEncoding) opt {
	return func(o *RSA) {
		o.coding = coding
	}
}

func New(opts ...opt) (os *ECB) {
	os = &ECB{
		padding: openssl.PKCS7_PADDING,
		coding:  std.New(),
	}
	for _, fn := range opts {
		fn(os)
	}
	return os
}

func (o *ECB) Encrypt(origin []byte) (encrypt string, err error) {
	var en []byte
	en, err = openssl.AesECBEncrypt(origin, o.key, o.padding)
	encrypt = o.coding.Encode(en)
	return
}

func (o *ECB) Decrypt(encrypt string) (origin []byte, err error) {
	var en []byte
	en, err = o.coding.Decode(encrypt)
	return openssl.AesECBDecrypt(en, o.key, o.padding)
}

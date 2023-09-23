package crypt

type ICrypt interface {
	Encrypt(origin []byte) (encrpy string, err error)
	Decrypt(encrpy string) (origin []byte, err error)
}

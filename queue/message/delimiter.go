package queue

/*
 * @abstract delimiter message
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import "strings"

type Delimiter struct {
	message   []byte
	delimiter string
}

func (m *Delimiter) Delimiter(delimiter string) *Delimiter {
	d.delimiter = delimiter
	return m
}

func (m *Delimiter) Fmt(key string) (b []byte) {
	return []byte(key + d.delimiter + string(d.msg))
}

// !!! key must be half-angle
func (m *Delimiter) Parse() (key string, value []byte) {
	if d.delimiter == "" {
		return "", d.msg
	}

	s := string(d.msg)
	if i := strings.Index(s, d.delimiter); i > 0 {
		key = s[:i]
		value = []byte(string([]rune(s)[i+1:]))
		return
	}

	return key, d.msg
}

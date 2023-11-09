package xml

import (
	"bytes"
	"encoding/xml"
	oxml "encoding/xml"
	"io"
	"io/ioutil"
	"regexp"

	"github.com/go-kratos/kratos/v2/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// Name is the name registered for the xml codec.
const Name = "xml"

func init() {
	encoding.RegisterCodec(Codec{})
}

// codec is a Codec implementation with xml.
type Codec struct{}

func (Codec) Marshal(v interface{}) ([]byte, error) {
	return xml.Marshal(v)
}

func (c Codec) Unmarshal(data []byte, v interface{}) (err error) {
	var xmlBytes []byte
	xmlBytes, err = c.GbkToUtf8(data)
	if err != nil {
		return
	}
	decoder := oxml.NewDecoder(bytes.NewReader(xmlBytes))
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return transform.NewReader(input, simplifiedchinese.GBK.NewEncoder()), nil
	}

	err = decoder.Decode(v)
	return
}

func (c Codec) Name() string {
	return Name
}

var regexpEncoding = regexp.MustCompile(`\sencoding=['"]?\w+['"]?`)

func (c Codec) GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}

	str := string(d)
	str = regexpEncoding.ReplaceAllString(str, " encoding='UTF-8'")

	return []byte(str), nil
}

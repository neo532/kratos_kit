package xml

import (
	"bytes"
	"encoding/xml"
	oxml "encoding/xml"
	"io"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-kratos/kratos/v2/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// Name is the name registered for the xml codec.
const Name = "xml"

func init() {
	encoding.RegisterCodec(Codec{})
}

var regexpEncoding = regexp.MustCompile(`\sencoding=['"]?([^'"]+)['"]?`)

// codec is a Codec implementation with xml.
type Codec struct{}

func (Codec) Marshal(v interface{}) ([]byte, error) {
	return xml.Marshal(v)
}

func (c Codec) Unmarshal(data []byte, v interface{}) (err error) {
	if s := string(data); len(s) > 40 {

		s, _ = url.PathUnescape(s)
		s = strings.ReplaceAll(s, "+", " ")

		if match := regexpEncoding.FindStringSubmatch(s[:40]); len(match) > 1 {
			switch strings.ToUpper(match[1]) {
			case "GBK", "GB2312":
				if data, err = c.GbkToUtf8([]byte(s)); err != nil {
					return
				}
				decoder := oxml.NewDecoder(bytes.NewReader(data))
				decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
					return transform.NewReader(input, simplifiedchinese.GBK.NewEncoder()), nil
				}

				err = decoder.Decode(v)
				return
			}
		}
	}
	err = xml.Unmarshal(data, v)
	return
}

func (c Codec) Name() string {
	return Name
}

func (c Codec) GbkToUtf8(s []byte) (byt []byte, err error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	if byt, err = ioutil.ReadAll(reader); err != nil {
		return
	}

	str := string(byt)
	str = regexpEncoding.ReplaceAllString(str, " encoding='UTF-8'")
	byt = []byte(str)
	return
}

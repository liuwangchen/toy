package verify

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/url"
	"sort"
	"strings"
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func Test_Sign(t *testing.T) {

	// [gift_info:[{"game_product":"{\"374\":[{\"product_id\":\"3\",\"product_name\":\"\\u94f6\\u4e24\",\"num\":\"100\"}]}"}] sign:[3999af3cf0395827c3719fe678b9fe63]]

	r := url.Values{
		`account`:   {`1185104511367119104`},
		`code`:      {`c_CJ5TVQ0HWT8`},
		`app_id`:    {`43BiUzHJ`},
		`game_id`:   {`374`},
		`gift_info`: {`{"game_product":{"374":[{"product_name":"劣质松纹铜","product_id":"1","num":"5"}]}}`},
		`op_id`:     {`2106`},
		`role_id`:   {`1185116889664192768`},
		`server_id`: {`1695420001`},
	}
	EncoderSign(r, "WDxG4oDqbVdstCjB")
	fmt.Println(r)
}

// convert UTF-8 to GBK
func EncodeGBK(s []byte) ([]byte, error) {
	I := bytes.NewReader(s)
	O := transform.NewReader(I, simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil, e
	}
	return d, nil
}

// Encode encodes the values into ``URL encoded'' form
// ("bar=baz&foo=quux") sorted by key.
func Encode(v url.Values) string {
	if v == nil {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		keyEscaped := url.QueryEscape(k)
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEscaped)
			buf.WriteByte('=')
			buf.WriteString(url.QueryEscape(v))
		}
	}
	return buf.String()
}

func Test_Sign1(t *testing.T) {

	// [gift_info:[{"game_product":"{\"374\":[{\"product_id\":\"3\",\"product_name\":\"\\u94f6\\u4e24\",\"num\":\"100\"}]}"}] sign:[3999af3cf0395827c3719fe678b9fe63]]

	r := url.Values{
		`app_id`:    {`2`},
		`qn_id`:     {`19`},
		`role_id`:   {`1203975211469242880`},
		`role_name`: {`1`},
		// `server_id`:   {`1000410002`},
		`server_id`:   {`1638410001`},
		`server_name`: {`1`},
		`timeline`:    {fmt.Sprintf("%d", 1575893823)},
	}
	EncoderSign(r, "A1i2RlFWBT3NPzQF")

	x := r.Encode()
	a, _ := url.QueryUnescape(x)
	b := base64.StdEncoding.EncodeToString([]byte(a))
	c := url.QueryEscape(b)
	d, _ := url.QueryUnescape(c)
	e, _ := base64.StdEncoding.DecodeString(d)
	f := string(e)
	fmt.Println(f)
}

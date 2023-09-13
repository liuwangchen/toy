package verify

import (
	"crypto/md5"
	"encoding/hex"
	"net/url"

	l4g "github.com/liuwangchen/toy/logger"
)

// VerifySign 校验
func VerifySign(r url.Values, key string, sign string) bool {

	decodeStr, err := url.QueryUnescape(r.Encode())
	if err != nil {
		return false
	}

	signStr := decodeStr + key
	waiteSign := md5.Sum([]byte(signStr))
	return hex.EncodeToString(waiteSign[:]) == sign
}

func CheckValues(r url.Values, params []string) bool {
	for _, param := range params {
		if "" == r.Get(param) {
			return false
		}
	}
	return true
}

// EncoderSign 编码签名
func EncoderSign(r url.Values, key string) {
	decodeStr, err := url.QueryUnescape(r.Encode())
	if err != nil {
		l4g.Error("[VerifySign] Error crypto url.QueryUnescape error=[%s]", err.Error())
	}

	signStr := decodeStr + key
	waiteSign := md5.Sum([]byte(signStr))
	r.Set("sign", hex.EncodeToString(waiteSign[:]))
}

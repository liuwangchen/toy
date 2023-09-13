package aesx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var key = "vip:customer:key"

func TestAes(t *testing.T) {
	testCases := []struct {
		origin string
	}{
		{"[1777310860,1777310950]"},
		{""},
		{"[1777310860,1777310950,1777310951]"},
		{"hello"},
	}
	s := "1234567890"
	for i := 0; i < 100; i++ {
		s += "1"
		testCases = append(testCases, struct {
			origin string
		}{
			s,
		})
	}

	for _, tc := range testCases {
		encryptCode, err := AesEncrypt(tc.origin, key)
		if err != nil {
			t.Error(err)
			return
		}
		decryptCode, err := AesDecrypt(encryptCode, key)
		if err != nil {
			t.Error(err)
			return
		}
		assert.EqualValues(t, tc.origin, decryptCode)
	}
}

// func FuzzAes(f *testing.F) {
// 	testCases := []struct {
// 		origin string
// 	}{
// 		{"[1777310860,1777310950]"},
// 		{""},
// 		{"[1777310860,1777310950,1777310951]"},
// 		{"hello"},
// 	}
//
// 	for _, tc := range testCases {
// 		f.Add(tc.origin)
// 	}
//
// 	f.Fuzz(func(t *testing.T, origin string) {
// 		encryptCode := AesEncrypt(origin, key)
// 		decryptCode := AesDecrypt(encryptCode, key)
// 		assert.EqualValues(t, origin, decryptCode)
// 	})
// }

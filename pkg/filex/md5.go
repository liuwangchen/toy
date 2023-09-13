package filex

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

// GetFileMD5 获取文件md5
func GetFileMD5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	} else {
		return hex.EncodeToString(h.Sum(nil)), nil
	}
}

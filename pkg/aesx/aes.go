package aesx

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
)

func AesEncrypt(orig string, key string) (val string, err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			err = fmt.Errorf("AesEncrypt err:%v", err1)
			return
		}
	}()
	// 转成字节数组
	origData := []byte(orig)
	k := []byte(key)
	// 分组秘钥
	// NewCipher该函数限制了输入k的长度必须为16, 24或者32
	block, err := aes.NewCipher(k)
	if err != nil {
		return "", err
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补全码
	origData, err = PKCS7Padding(origData, blockSize)
	if err != nil {
		return "", err
	}
	// 加密模式
	blockMode := cipher.NewCBCEncrypter(block, k[:blockSize])
	// 创建数组
	cryted := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(cryted, origData)
	return base64.StdEncoding.EncodeToString(cryted), nil
}

func AesDecrypt(cryted string, key string) (val string, err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			err = fmt.Errorf("AesDecrypt err:%v", err1)
			return
		}
	}()
	// 转成字节数组
	crytedByte, _ := base64.StdEncoding.DecodeString(cryted)
	k := []byte(key)
	// 分组秘钥
	block, err := aes.NewCipher(k)
	if err != nil {
		return "", err
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 加密模式
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
	// 创建数组
	orig := make([]byte, len(crytedByte))
	// 解密
	blockMode.CryptBlocks(orig, crytedByte)
	// 去补全码
	orig, err = PKCS7UnPadding(orig)
	if err != nil {
		return "", err
	}
	return string(orig), nil
}

// PKCS7Padding 补码
// AES加密数据块分组长度必须为128bit(byte[16])，密钥长度可以是128bit(byte[16])、192bit(byte[24])、256bit(byte[32])中的任意一个。
func PKCS7Padding(ciphertext []byte, blocksize int) (val []byte, err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			err = fmt.Errorf("PKCS7Padding err:%v", err1)
			return
		}
	}()
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...), nil
}

// PKCS7UnPadding 去码
func PKCS7UnPadding(origData []byte) (val []byte, err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			err = fmt.Errorf("PKCS7UnPadding err:%v", err1)
			return
		}
	}()
	if len(origData) == 0 {
		return nil, errors.New("加密字符串错误")
	}
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)], nil
}

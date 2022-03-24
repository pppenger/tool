package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
)

type PhoneSecret struct {
	Phone       string
	SecretPhone string
}

func isPlainPhone(phone string) (bool, error) {
	isPlain := true
	for _, ch := range phone {
		if ch >= 'a' && ch < 'g' {
			isPlain = false
			continue
		}

		if ch >= '0' && ch <= '9' {
			continue
		}

		return isPlain, fmt.Errorf("phone %v is invalid", phone)
	}
	return isPlain, nil
}

const sourceAesKey = "xiangyuxingqiu_user_account_base"

func NewPhoneSecret(phone string) (*PhoneSecret, error) {
	isPlain, err := isPlainPhone(phone)
	if err != nil {
		return nil, err
	}

	secret := &PhoneSecret{}
	if isPlain {
		secret.Phone = phone
		var bphone = []byte(phone)
		var bkey = []byte(sourceAesKey)
		srca := EncryptAES(bphone, bkey)
		hexsrc := hex.EncodeToString(srca)
		secret.SecretPhone = hexsrc
		return secret, nil
	}

	var hexdst []byte
	hexdst, _ = hex.DecodeString(phone)
	var bkey = []byte(sourceAesKey)
	desa := DecryptAES(hexdst, bkey)
	secret.Phone = string(desa[:])
	secret.SecretPhone = phone

	return secret, nil
}

//aes解密
func DecryptAES(src, key []byte) []byte {
	//1.创建并返回一个使用DES算法的cipher.Block接口。
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	//2.crypto解密
	blockModel := cipher.NewCBCDecrypter(block, key[:block.BlockSize()]) //block.BlockSize() ==len(key)
	//3.解密连续块
	blockModel.CryptBlocks(src, src)
	//.删除填充数组
	src = DePadding(src)

	return src
}
func DePadding(src []byte) []byte {
	//1.取出最后一个元素
	lastNum := int(src[len(src)-1])
	//2.删除和最后一个元素相等长的字节
	newText := src[:len(src)-lastNum]
	return newText
}

func EncryptAES(src, key []byte) []byte {
	//1.创建并返回一个使用DES算法的cipher.Block接口。
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	//2.对src进行填充
	src = Padding(src, block.BlockSize())
	blockModel := cipher.NewCBCEncrypter(block, key[:block.BlockSize()]) //block.BlockSize() ==len(key)
	//4.crypto加密连续块
	blockModel.CryptBlocks(src, src)

	return src
}
func Padding(src []byte, blockSize int) []byte {
	//func padding(src []byte, blockSize int) {
	//1.截取加密代码 段数
	padding := blockSize - len(src)%blockSize
	//2.有余数
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	//3.添加余数
	src = append(src, padText...)
	return src
}

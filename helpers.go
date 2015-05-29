package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
)

func randStr(strSize int) string {

	var dictionary string
	dictionary = "0123456789abcdefghijklmnopqrstuvwxyz"

	bytes := GenerateIv(strSize)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}

func Encrypt(msg, key string, iv []byte) (string, error) {
	src := []byte(msg)
	dst := make([]byte, len(src))

	aesBlockEncrypter, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	aesEncrypter := cipher.NewCFBEncrypter(aesBlockEncrypter, iv)
	aesEncrypter.XORKeyStream(dst, src)

	return base64.StdEncoding.EncodeToString(dst), nil
}

func Decrypt(msg, key string, iv []byte) (string, error) {
	src, _ := base64.StdEncoding.DecodeString(msg)
	dst := make([]byte, len(src))

	aesBlockDecrypter, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	aesDecrypter := cipher.NewCFBDecrypter(aesBlockDecrypter, iv)
	aesDecrypter.XORKeyStream(dst, src)

	return string(dst), nil
}

func GenerateIv(size int) []byte {
	var bytes = make([]byte, size)
	rand.Read(bytes)

	return bytes
}

func GenerateSalt(size int) string {
	iv := GenerateIv(size)

	return base64.StdEncoding.EncodeToString(iv)
}

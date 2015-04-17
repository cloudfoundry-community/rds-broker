package main

import (
	"crypto/rand"
)

func randStr(strSize int) string {

	var dictionary string
	dictionary = "0123456789abcdefghijklmnopqrstuvwxyz"

	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}

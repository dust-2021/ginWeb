package main

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

func curve25519Test() {
	// 生成随机私钥
	var privateKey [32]byte
	_, err := rand.Read(privateKey[:])
	if err != nil {
		panic(fmt.Errorf("generate wg key error"))
	}
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	publicKey, err := curve25519.X25519(privateKey[:], curve25519.Basepoint)
	fmt.Printf("pub: %v \r\n pri: %v\r\n", publicKey, privateKey)
}

func main() {
	curve25519Test()
}

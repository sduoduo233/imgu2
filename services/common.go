package services

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
)

func RandomHexString(n int) string {
	buf := make([]byte, n)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(buf)
}

func RandomNumber(min, max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	if err != nil {
		panic(err)
	}

	return int(n.Int64()) + min
}

func RandomString(length int) string {
	s := ""

	chars := "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

	for i := 0; i < length; i++ {
		s += string(chars[RandomNumber(0, len(chars))])
	}

	return s
}

package utils

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
)

func RandomString(n int) string {
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

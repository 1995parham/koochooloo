package model

import (
	"crypto/rand"
	"log"
	"math/big"
)

// Length is a random key length.
const Length = 6

const source = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

// Key generates a random key from the source characters.
func Key() string {
	b := make([]byte, Length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(source))))
		if err != nil {
			log.Fatal(err)
		}

		b[i] = source[n.Int64()]
	}

	return string(b)
}

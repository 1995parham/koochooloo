package generator

import (
	"crypto/rand"
	"math/big"
)

type Simple struct{}

// ShortURLKey generates a random key from the source characters.
func (Simple) ShortURLKey() string {
	const (
		length = 6
		source = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	)

	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(source))))
		if err != nil {
			panic(err)
		}

		b[i] = source[n.Int64()]
	}

	return string(b)
}

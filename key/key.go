package key

import "math/rand"

// Length is a random key length
const Length = 6

const source = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

// Key generates a random key from the source
func Key() string {
	b := make([]byte, Length)
	for i := range b {
		b[i] = source[rand.Intn(len(source))]
	}
	return string(b)
}

package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/constellation39/majsoul/message"
	"math/rand"
)

// HashPassword It returns the hex encoded string of the HMAC.
func HashPassword(password string) string {
	hash := hmac.New(sha256.New, []byte("lailai"))
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

var keys = []int{0x84, 0x5e, 0x4e, 0x42, 0x39, 0xa2, 0x1f, 0x60, 0x1c}

// DecodeActionPrototype modifies the Data field of a given
func DecodeActionPrototype(actionPrototype *message.ActionPrototype) {
	for i := 0; i < len(actionPrototype.Data); i++ {
		u := (23 ^ len(actionPrototype.Data)) + 5*i + keys[i%len(keys)]&255
		actionPrototype.Data[i] ^= byte(u)
	}
}

// UUID generates a pseudo-random UUID-like string.
func UUID() string {
	const charSet = "0123456789abcdefghijklmnopqrstuvwxyz"
	csl := len(charSet)
	b := make([]byte, 36)
	for i := 0; i < 36; i++ {
		if i == 7 || i == 12 || i == 17 || i == 22 {
			b[i] = '-'
			continue
		}
		b[i] = charSet[rand.Intn(csl)]
	}
	return string(b)
}

package majsoul

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// hashPassword password with hmac sha256
// return hash string
func hashPassword(data string) string {
	hash := hmac.New(sha256.New, []byte("lailai"))
	hash.Write([]byte(data))
	return hex.EncodeToString(hash.Sum(nil))
}

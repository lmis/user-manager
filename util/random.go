package util

import (
	"crypto/rand"
	b64 "encoding/base64"
)

func MakeRandomURLSafeB64(size int) string {
	randomBytes := make([]byte, size)

	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return b64.URLEncoding.EncodeToString(randomBytes)

}

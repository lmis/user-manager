package util

import (
	"crypto/rand"
	b64 "encoding/base64"
)

func MakeRandomURLSafeB64(size int) (string, error) {
	res := ""
	randomBytes := make([]byte, size)

	_, err := rand.Read(randomBytes)
	if err == nil {
		res = b64.URLEncoding.EncodeToString(randomBytes)
	}

	return res, err
}

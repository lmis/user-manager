package random

import (
	"crypto/rand"
	"encoding/base64"
	"user-manager/util/errs"
)

func MakeRandomURLSafeB64(size int) string {
	randomBytes := make([]byte, size)

	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(errs.Wrap("issue reading random bytes", err))
	}
	return base64.URLEncoding.EncodeToString(randomBytes)
}

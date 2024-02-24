package service

import (
	"crypto/rand"
	dm "user-manager/domain-model"
	"user-manager/util/errs"

	"golang.org/x/crypto/argon2"
)

func MakeCredentials(password []byte) (dm.UserCredentials, error) {
	if len(password) < 8 {
		return dm.UserCredentials{}, errs.Error("password too short")
	}
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return dm.UserCredentials{}, errs.Wrap("error generating salt", err)
	}

	// See https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#input-limits
	key := argon2.IDKey(password, salt, 1, 64*1024, 4, 32)

	return dm.UserCredentials{Key: key, Salt: salt}, nil
}

func VerifyCredentials(password []byte, credentials dm.UserCredentials) bool {
	key := argon2.IDKey(password, credentials.Salt, 1, 64*1024, 4, 32)
	return string(key) == string(credentials.Key)
}

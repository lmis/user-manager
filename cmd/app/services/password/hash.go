package password

import (
	"golang.org/x/crypto/bcrypt"
)

func Hash(password []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

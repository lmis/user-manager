package service

import (
	"user-manager/util/errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
}

func ProvideAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Hash(password []byte) (string, error) {
	if len(password) < 8 {
		return "", errors.Error("password too short")

	}
	// See https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#input-limits
	if len(password) > 71 {
		return "", errors.Error("password too long")

	}
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

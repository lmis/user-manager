package service

import (
	"user-manager/util"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
}

func ProvideAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Hash(password []byte) (string, error) {
	if len(password) > 71 {
		return "", util.Errorf("password too long")

	}
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

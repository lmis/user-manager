package service

import "golang.org/x/crypto/bcrypt"

type AuthService struct {
}

func ProvideAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Hash(password []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

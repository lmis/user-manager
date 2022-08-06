package settings

import (
	ginext "user-manager/cmd/app/gin-extensions"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type SudoTO struct {
	Password []byte `json:"password"`
}

type SudoResponseTO struct {
	Success bool `json:"success"`
}

func PostSudo(requestContext *ginext.RequestContext, requestTO *SudoTO, _ *gin.Context) (*SudoResponseTO, error) {
	securityLog := requestContext.SecurityLog
	user := requestContext.Authentication.AppUser

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), requestTO.Password); err != nil {
		securityLog.Info("Password mismatch in sudo attempt for user %s", user.AppUserID)
		return &SudoResponseTO{}, nil
	}
	return &SudoResponseTO{Success: true}, nil
}

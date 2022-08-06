package settings

import (
	ginext "user-manager/cmd/app/gin-extensions"
	user_service "user-manager/cmd/app/services/user"
	"user-manager/db/generated/models"
	"user-manager/util"

	"github.com/gin-gonic/gin"
)

type LanguageTO struct {
	Language models.UserLanguage `json:"language"`
}

func PostLanguage(requestContext *ginext.RequestContext, requestTO *LanguageTO, _ *gin.Context) error {
	user := requestContext.Authentication.AppUser

	user.Language = requestTO.Language

	if err := user_service.UpdateUser(requestContext, user); err != nil {
		return util.Wrap("error updating language", err)
	}
	return nil
}

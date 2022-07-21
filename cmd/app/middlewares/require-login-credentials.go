package middleware

import (
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	auth_service "user-manager/cmd/app/services/auth"
	"user-manager/util"

	"github.com/gin-gonic/gin"
)

type LoginCredentialsTO struct {
	Email        string `json:"email"`
	Password     []byte `json:"password"`
	SecondFactor string `json:"secondFactor"`
}

func RequireLoginCredentials(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	authentication := requestContext.Authentication

	var loginCredentialsTO LoginCredentialsTO
	if err := c.BindJSON(&loginCredentialsTO); err != nil {
		c.AbortWithError(http.StatusBadRequest, util.Wrap("cannot bind to loginCredentialsTO", err))
		return
	}

	if err := auth_service.ValidateLogin(
		requestContext,
		authentication.AppUser,
		loginCredentialsTO.Password,
		loginCredentialsTO.SecondFactor,
	); err != nil {
		status := http.StatusInternalServerError
		if _, ok := err.(*auth_service.ValidateLoginError); ok {
			status = http.StatusBadRequest
		}
		c.AbortWithError(status, util.Wrap("issue validating login credentials", err))
	}
}

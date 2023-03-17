// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package resource

import (
	"github.com/gin-gonic/gin"
	"user-manager/cmd/app/injector"
	"user-manager/repository"
	"user-manager/service"
)

// Injectors from wire.go:

func InitializeEmailConfirmationResource(c *gin.Context) *EmailConfirmationResource {
	securityLog := injector.ProvideSecurityLog(c)
	tx := injector.ProvideTx(c)
	mailQueueRepository := repository.ProvideMailQueueRepository(tx)
	config := injector.ProvideConfig()
	v := injector.ProvideTranslations()
	template := injector.ProvideBaseTemplate()
	mailQueueService := service.ProvideMailQueueService(mailQueueRepository, config, v, template)
	userSession := injector.ProvideUserSession(c)
	userRepository := repository.ProvideUserRepository(tx)
	emailConfirmationResource := ProvideEmailConfirmationResource(securityLog, mailQueueService, userSession, userRepository)
	return emailConfirmationResource
}

func InitializeLoginResource(c *gin.Context) *LoginResource {
	securityLog := injector.ProvideSecurityLog(c)
	config := injector.ProvideConfig()
	sessionCookieService := service.ProvideSessionCookieService(c, config)
	tx := injector.ProvideTx(c)
	sessionRepository := repository.ProvideSessionRepository(tx)
	userRepository := repository.ProvideUserRepository(tx)
	secondFactorThrottlingRepository := repository.ProvideSecondFactorThrottlingRepository(tx)
	loginResource := ProvideLoginResource(securityLog, sessionCookieService, sessionRepository, userRepository, secondFactorThrottlingRepository)
	return loginResource
}

func InitializeSignUpResource(c *gin.Context) *SignUpResource {
	tx := injector.ProvideTx(c)
	userRepository := repository.ProvideUserRepository(tx)
	mailQueueRepository := repository.ProvideMailQueueRepository(tx)
	config := injector.ProvideConfig()
	v := injector.ProvideTranslations()
	template := injector.ProvideBaseTemplate()
	mailQueueService := service.ProvideMailQueueService(mailQueueRepository, config, v, template)
	authService := service.ProvideAuthService()
	securityLog := injector.ProvideSecurityLog(c)
	signUpResource := ProvideSignUpResource(userRepository, mailQueueService, authService, securityLog)
	return signUpResource
}

func InitializeUserInfoResource(c *gin.Context) *UserInfoResource {
	userSession := injector.ProvideUserSession(c)
	userInfoResource := ProvideUserInfoResource(userSession)
	return userInfoResource
}

func InitializeLogoutResource(c *gin.Context) *LogoutResource {
	securityLog := injector.ProvideSecurityLog(c)
	config := injector.ProvideConfig()
	sessionCookieService := service.ProvideSessionCookieService(c, config)
	tx := injector.ProvideTx(c)
	sessionRepository := repository.ProvideSessionRepository(tx)
	userSession := injector.ProvideUserSession(c)
	logoutResource := ProvideLogoutResource(securityLog, sessionCookieService, sessionRepository, userSession)
	return logoutResource
}

func InitializeSettingsResource(c *gin.Context) *SettingsResource {
	securityLog := injector.ProvideSecurityLog(c)
	config := injector.ProvideConfig()
	sessionCookieService := service.ProvideSessionCookieService(c, config)
	userSession := injector.ProvideUserSession(c)
	tx := injector.ProvideTx(c)
	userRepository := repository.ProvideUserRepository(tx)
	sessionRepository := repository.ProvideSessionRepository(tx)
	settingsResource := ProvideSettingsResource(securityLog, sessionCookieService, userSession, userRepository, sessionRepository)
	return settingsResource
}

func InitializeSensitiveSettingsResource(c *gin.Context) *SensitiveSettingsResource {
	securityLog := injector.ProvideSecurityLog(c)
	tx := injector.ProvideTx(c)
	mailQueueRepository := repository.ProvideMailQueueRepository(tx)
	config := injector.ProvideConfig()
	v := injector.ProvideTranslations()
	template := injector.ProvideBaseTemplate()
	mailQueueService := service.ProvideMailQueueService(mailQueueRepository, config, v, template)
	userSession := injector.ProvideUserSession(c)
	userRepository := repository.ProvideUserRepository(tx)
	sensitiveSettingsResource := ProvideSensitiveSettingsResource(securityLog, mailQueueService, userSession, userRepository)
	return sensitiveSettingsResource
}

func InitializeResetPasswordResource(c *gin.Context) *ResetPasswordResource {
	securityLog := injector.ProvideSecurityLog(c)
	authService := service.ProvideAuthService()
	tx := injector.ProvideTx(c)
	mailQueueRepository := repository.ProvideMailQueueRepository(tx)
	config := injector.ProvideConfig()
	v := injector.ProvideTranslations()
	template := injector.ProvideBaseTemplate()
	mailQueueService := service.ProvideMailQueueService(mailQueueRepository, config, v, template)
	userRepository := repository.ProvideUserRepository(tx)
	resetPasswordResource := ProvideResetPasswordResource(securityLog, authService, mailQueueService, userRepository)
	return resetPasswordResource
}

package resource

import (
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util/errors"

	"github.com/gin-gonic/gin"
)

type LogoutResource struct {
	securityLog          dm.SecurityLog
	sessionCookieService *service.SessionCookieService
	sessionRepository    *repository.SessionRepository
	userSession          dm.UserSession
}

func ProvideLogoutResource(
	securityLog dm.SecurityLog,
	sessionCookieService *service.SessionCookieService,
	sessionRepository *repository.SessionRepository,
	userSession dm.UserSession,
) *LogoutResource {
	return &LogoutResource{securityLog, sessionCookieService, sessionRepository, userSession}
}

func RegisterLogoutResource(group *gin.RouterGroup) {
	group.POST("logout", ginext.WrapEndpointWithoutResponseBody(InitializeLogoutResource, (*LogoutResource).Logout))
}

type LogoutTO struct {
	ForgetDevice bool `json:"forgetDevice"`
}

func (r *LogoutResource) Logout(request LogoutTO) error {
	securityLog := r.securityLog
	userSession := r.userSession
	sessionCookieService := r.sessionCookieService
	sessionRepository := r.sessionRepository

	securityLog.Info("Logout")

	sessionCookieService.RemoveSessionCookie(dm.UserSessionTypeLogin)
	if userSession.UserSessionID != "" {
		if err := sessionRepository.Delete(userSession.UserSessionID); err != nil {
			return errors.Wrap("issue while deleting login session", err)
		}
	}

	sudoSessionID, err := sessionCookieService.GetSessionCookie(dm.UserSessionTypeSudo)
	if err != nil {
		return errors.Wrap("issue reading sudo session cookie", err)
	}
	if sudoSessionID != "" {
		sessionCookieService.RemoveSessionCookie(dm.UserSessionTypeSudo)
		if err := sessionRepository.Delete(sudoSessionID); err != nil {
			return errors.Wrap("issue while deleting sudo session", err)
		}
	}
	if request.ForgetDevice {
		deviceSessionID, err := sessionCookieService.GetSessionCookie(dm.UserSessionTypeRememberDevice)
		if err != nil {
			return errors.Wrap("issue reading device session cookie", err)
		}
		if deviceSessionID != "" {
			sessionCookieService.RemoveSessionCookie(dm.UserSessionTypeRememberDevice)
			if err := sessionRepository.Delete(deviceSessionID); err != nil {
				return errors.Wrap("issue while deleting device session", err)
			}
		}
	}

	return nil
}

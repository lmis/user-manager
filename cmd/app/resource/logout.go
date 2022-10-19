package resource

import (
	ginext "user-manager/cmd/app/gin-extensions"
	domain_model "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util"
	"user-manager/util/nullable"

	"github.com/gin-gonic/gin"
)

type LogoutResource struct {
	securityLog          domain_model.SecurityLog
	sessionCookieService *service.SessionCookieService
	sessionRepository    *repository.SessionRepository
	userSession          nullable.Nullable[*domain_model.UserSession]
}

func ProvideLogoutResource(
	securityLog domain_model.SecurityLog,
	sessionCookieService *service.SessionCookieService,
	sessionRepository *repository.SessionRepository,
	userSession nullable.Nullable[*domain_model.UserSession],
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

	sessionCookieService.RemoveSessionCookie(domain_model.USER_SESSION_TYPE_LOGIN)
	if userSession.IsPresent {
		if err := sessionRepository.Delete(userSession.OrPanic().UserSessionID); err != nil {
			return util.Wrap("issue while deleting login session", err)
		}
	}

	sudoSessionId, err := sessionCookieService.GetSessionCookie(domain_model.USER_SESSION_TYPE_SUDO)
	if err != nil {
		return util.Wrap("issue reading sudo session cookie", err)
	}
	if sudoSessionId.IsPresent {
		sessionCookieService.RemoveSessionCookie(domain_model.USER_SESSION_TYPE_SUDO)
		if err := sessionRepository.Delete(domain_model.UserSessionID(sudoSessionId.OrPanic())); err != nil {
			return util.Wrap("issue while deleting sudo session", err)
		}
	}
	if request.ForgetDevice {
		deviceSessionId, err := sessionCookieService.GetSessionCookie(domain_model.USER_SESSION_TYPE_REMEMBER_DEVICE)
		if err != nil {
			return util.Wrap("issue reading device session cookie", err)
		}
		if deviceSessionId.IsPresent {
			sessionCookieService.RemoveSessionCookie(domain_model.USER_SESSION_TYPE_REMEMBER_DEVICE)
			if err := sessionRepository.Delete(domain_model.UserSessionID(deviceSessionId.OrPanic())); err != nil {
				return util.Wrap("issue while deleting device session", err)
			}
		}
	}

	return nil
}
package domain_model

import (
	"user-manager/db/generated/models"
)

type UserSessionID string
type UserSessionType models.UserSessionType

const (
	USER_SESSION_TYPE_LOGIN           = UserSessionType(models.UserSessionTypeLOGIN)
	USER_SESSION_TYPE_SUDO            = UserSessionType(models.UserSessionTypeSUDO)
	USER_SESSION_TYPE_REMEMBER_DEVICE = UserSessionType(models.UserSessionTypeREMEMBER_DEVICE)
)

type UserSession struct {
	UserSessionID   UserSessionID   `json:"userSessionId"`
	User            *AppUser        `json:"user"`
	UserSessionType UserSessionType `json:"userSessionType"`
}

func FromUserSessionAppUserAndUserRolesModel(m *models.UserSession, u *models.AppUser, r models.AppUserRoleSlice) *UserSession {
	return &UserSession{
		UserSessionID:   UserSessionID(m.UserSessionID),
		User:            FromAppUserAndUserRolesModel(u, r),
		UserSessionType: UserSessionType(m.UserSessionType),
	}
}

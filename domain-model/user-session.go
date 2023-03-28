package domain_model

import (
	"user-manager/db/generated/models/postgres/public/model"

	"github.com/go-jet/jet/v2/postgres"
)

type UserSessionID string
type UserSessionType model.UserSessionType

const (
	UserSessionTypeLogin          = UserSessionType(model.UserSessionType_Login)
	UserSessionTypeSudo           = UserSessionType(model.UserSessionType_Sudo)
	UserSessionTypeRememberDevice = UserSessionType(model.UserSessionType_RememberDevice)
)

type UserSession struct {
	UserSessionID   UserSessionID
	User            *AppUser        `json:"user"`
	UserSessionType UserSessionType `json:"userSessionType"`
}

func (u UserSessionType) String() string {
	return model.UserSessionType(u).String()
}

func (u UserSessionType) ToStringExpression() postgres.StringExpression {
	return postgres.NewEnumValue(model.UserSessionType(u).String())
}

func (u UserSessionID) ToStringExpression() postgres.StringExpression {
	return postgres.String(string(u))
}

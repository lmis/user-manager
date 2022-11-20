package domain_model

import (
	"encoding/json"
	"time"
	"user-manager/util/nullable"

	"github.com/go-jet/jet/v2/postgres"
)

type AppUserID int64

type AppUser struct {
	AppUserID                    AppUserID
	Language                     UserLanguage
	UserName                     string
	PasswordHash                 string
	Email                        string
	EmailVerified                bool
	EmailVerificationToken       nullable.Nullable[string]
	NextEmail                    nullable.Nullable[string]
	PasswordResetToken           nullable.Nullable[string]
	PasswordResetTokenValidUntil nullable.Nullable[time.Time]
	SecondFactorToken            nullable.Nullable[string]
	TemporarySecondFactorToken   nullable.Nullable[string]
	UserRoles                    []UserRole
}

func (u *AppUser) MarshallJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":        u.AppUserID,
		"userRoles": u.UserRoles,
	})
}

func (a AppUserID) ToIntegerExpression() postgres.IntegerExpression {
	return postgres.Int64(int64(a))
}

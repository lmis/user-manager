package domain_model

import (
	"encoding/json"
	"time"
	"user-manager/db/generated/models"
	"user-manager/util/nullable"
	"user-manager/util/slices"
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

func FromAppUserAndUserRolesModel(m *models.AppUser, r models.AppUserRoleSlice) *AppUser {
	return &AppUser{
		AppUserID:                    AppUserID(m.AppUserID),
		Language:                     UserLanguage(m.Language),
		UserName:                     m.UserName,
		PasswordHash:                 m.PasswordHash,
		Email:                        m.Email,
		EmailVerified:                m.EmailVerified,
		EmailVerificationToken:       nullable.FromNullString(m.EmailVerificationToken),
		NextEmail:                    nullable.FromNullString(m.NextEmail),
		PasswordResetToken:           nullable.FromNullString(m.PasswordResetToken),
		PasswordResetTokenValidUntil: nullable.FromNullTime(m.PasswordResetTokenValidUntil),
		SecondFactorToken:            nullable.FromNullString(m.SecondFactorToken),
		TemporarySecondFactorToken:   nullable.FromNullString(m.TemporarySecondFactorToken),
		UserRoles:                    slices.Map(r, func(m *models.AppUserRole) UserRole { return UserRole(m.Role) }),
	}
}

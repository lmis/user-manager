package domain_model

import (
	"time"
	"user-manager/db/generated/models"
	"user-manager/util/nullable"
	"user-manager/util/slices"
)

type AppUserID int64

type AppUser struct {
	AppUserID                    AppUserID                    `json:"appUserId"`
	Language                     UserLanguage                 `json:"language"`
	UserName                     string                       `json:"userName"`
	PasswordHash                 string                       `json:"passwordHash"`
	Email                        string                       `json:"email"`
	EmailVerified                bool                         `json:"emailVerified"`
	EmailVerificationToken       nullable.Nullable[string]    `json:"emailVerificationToken,omitempty"`
	NewEmail                     nullable.Nullable[string]    `json:"newEmail,omitempty"`
	PasswordResetToken           nullable.Nullable[string]    `json:"passwordResetToken,omitempty"`
	PasswordResetTokenValidUntil nullable.Nullable[time.Time] `json:"passwordResetTokenValidUntil,omitempty"`
	TwoFactorToken               nullable.Nullable[string]    `json:"twoFactorToken,omitempty"`
	TempraryTwoFactorToken       nullable.Nullable[string]    `json:"tempraryTwoFactorToken,omitempty"`
	UserRoles                    []UserRole                   `json:"userRoles,omitempty"`
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
		NewEmail:                     nullable.FromNullString(m.NewEmail),
		PasswordResetToken:           nullable.FromNullString(m.PasswordResetToken),
		PasswordResetTokenValidUntil: nullable.FromNullTime(m.PasswordResetTokenValidUntil),
		TwoFactorToken:               nullable.FromNullString(m.TwoFactorToken),
		TempraryTwoFactorToken:       nullable.FromNullString(m.TempraryTwoFactorToken),
		UserRoles:                    slices.Map(r, func(m *models.AppUserRole) UserRole { return UserRole(m.Role) }),
	}
}

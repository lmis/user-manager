package domain_model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"slices"
	"time"
)

type UserID primitive.ObjectID

func (u UserID) ToString() string {
	return primitive.ObjectID(u).String()
}

type UserSessionType string
type UserSessionToken string
type UserLanguage string
type UserRole string

const (
	UserSessionTypeLogin          UserSessionType = "login"
	UserSessionTypeSudo           UserSessionType = "sudo"
	UserSessionTypeRememberDevice UserSessionType = "remember-device"

	UserLanguageEn UserLanguage = "en"
	UserLanguageDe UserLanguage = "de"

	UserRoleUser       UserRole = "user"
	UserRoleAdmin      UserRole = "admin"
	UserRoleSuperAdmin UserRole = "super-admin"

	UserCollectionName = "User"
)

func (l UserLanguage) IsValid() bool {
	return slices.Contains(AllUserLanguages(), l)
}

func AllUserLanguages() []UserLanguage {
	return []UserLanguage{UserLanguageEn, UserLanguageDe}
}

type SecondFactorThrottling struct {
	FailedAttemptsSinceLastSuccess int32
	TimeoutUntil                   time.Time
}
type UserSession struct {
	Token     UserSessionToken
	Type      UserSessionType
	TimeoutAt time.Time
}

func (u UserSession) IsPresent() bool {
	return u.Token != ""
}

type UserCredentials struct {
	Key  []byte
	Salt []byte
}

type User struct {
	ID                           UserID `bson:"_id,omitempty"`
	Language                     UserLanguage
	UserName                     string
	Credentials                  UserCredentials
	Email                        string
	EmailVerified                bool
	EmailVerificationToken       string
	NextEmail                    string
	PasswordResetToken           string
	PasswordResetTokenValidUntil time.Time
	SecondFactorToken            string
	TemporarySecondFactorToken   string
	UserRoles                    []UserRole
	Sessions                     []UserSession
	SecondFactorThrottling       SecondFactorThrottling
}

func (u User) IsPresent() bool {
	return u.ID != UserID(primitive.NilObjectID)
}

type UserInsert struct {
	Language                     UserLanguage
	UserName                     string
	Credentials                  UserCredentials
	Email                        string
	EmailVerified                bool
	EmailVerificationToken       string
	NextEmail                    string
	PasswordResetToken           string
	PasswordResetTokenValidUntil time.Time
	SecondFactorToken            string
	TemporarySecondFactorToken   string
	UserRoles                    []UserRole
	Sessions                     []UserSession
	SecondFactorThrottling       SecondFactorThrottling
}

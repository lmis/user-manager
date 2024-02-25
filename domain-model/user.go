package domain_model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type UserID primitive.ObjectID

type UserSessionType string
type UserSessionToken string
type UserRole string

const (
	UserSessionTypeLogin          UserSessionType = "LOGIN"
	UserSessionTypeSudo           UserSessionType = "SUDO"
	UserSessionTypeRememberDevice UserSessionType = "REMEMBER-DEVICE"

	UserRoleUser       UserRole = "user"
	UserRoleAdmin      UserRole = "admin"
	UserRoleSuperAdmin UserRole = "super-admin"

	UserCollectionName = "users"
)

type SecondFactorThrottling struct {
	FailuresSinceSuccess int32     `bson:"failedAttemptsSinceLastSuccess,omitempty"`
	TimeoutUntil         time.Time `bson:"timeoutUntil,omitempty"`
}
type UserSession struct {
	Token     UserSessionToken `bson:"token,omitempty"`
	Type      UserSessionType  `bson:"type,omitempty"`
	TimeoutAt time.Time        `bson:"timeoutAt,omitempty"`
}

func (u UserSession) IsPresent() bool {
	return u.Token != ""
}

type UserCredentials struct {
	Key  []byte
	Salt []byte
}

type User struct {
	ObjectID                     primitive.ObjectID     `bson:"_id,omitempty"`
	Name                         string                 `bson:"name,omitempty"`
	Credentials                  UserCredentials        `bson:"credentials,omitempty"`
	Email                        string                 `bson:"email,omitempty"`
	EmailVerified                bool                   `bson:"emailVerified,omitempty"`
	EmailVerificationToken       string                 `bson:"emailVerificationToken,omitempty"`
	NextEmail                    string                 `bson:"nextEmail,omitempty"`
	PasswordResetToken           string                 `bson:"passwordResetToken,omitempty"`
	PasswordResetTokenValidUntil time.Time              `bson:"passwordResetTokenValidUntil,omitempty"`
	SecondFactorToken            string                 `bson:"secondFactorToken,omitempty"`
	TemporarySecondFactorToken   string                 `bson:"temporarySecondFactorToken,omitempty"`
	UserRoles                    []UserRole             `bson:"userRoles,omitempty"`
	Sessions                     []UserSession          `bson:"sessions,omitempty"`
	SecondFactorThrottling       SecondFactorThrottling `bson:"secondFactorThrottling,omitempty"`
}

func (u User) ID() UserID {
	return UserID(u.ObjectID)
}

func (u User) IsPresent() bool {
	return u.ObjectID != primitive.NilObjectID
}

type UserInsert struct {
	UserName               string
	Credentials            UserCredentials
	Email                  string
	EmailVerified          bool
	EmailVerificationToken string
	UserRoles              []UserRole
}

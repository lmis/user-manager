package domain_model

import (
	"time"
	"user-manager/db/generated/models"
	"user-manager/util/nullable"
)

type TwoFactorThrottlingID int64

type SecondFactorThrottling struct {
	TwoFactorThrottlingID          TwoFactorThrottlingID        `json:"twoFactorThrottlingId"`
	AppUserID                      AppUserID                    `json:"appUserId"`
	FailedAttemptsSinceLastSuccess int                          `json:"failedAttemptsSinceLastSuccess"`
	TimeoutUntil                   nullable.Nullable[time.Time] `json:"timeoutUntil,omitempty"`
}

func FromSecondFactorThrottlingModel(m *models.TwoFactorThrottling) *SecondFactorThrottling {
	return &SecondFactorThrottling{
		TwoFactorThrottlingID:          TwoFactorThrottlingID(m.TwoFactorThrottlingID),
		AppUserID:                      AppUserID(m.AppUserID),
		FailedAttemptsSinceLastSuccess: m.FailedAttemptsSinceLastSuccess,
		TimeoutUntil:                   nullable.FromNullTime(m.TimeoutUntil),
	}
}

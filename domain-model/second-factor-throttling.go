package domain_model

import (
	"time"
	"user-manager/db/generated/models"
	"user-manager/util/nullable"
)

type SecondFactorThrottlingID int64

type SecondFactorThrottling struct {
	SecondFactorThrottlingID       SecondFactorThrottlingID     `json:"secondFactorThrottlingId"`
	AppUserID                      AppUserID                    `json:"appUserId"`
	FailedAttemptsSinceLastSuccess int                          `json:"failedAttemptsSinceLastSuccess"`
	TimeoutUntil                   nullable.Nullable[time.Time] `json:"timeoutUntil,omitempty"`
}

func FromSecondFactorThrottlingModel(m *models.SecondFactorThrottling) *SecondFactorThrottling {
	return &SecondFactorThrottling{
		SecondFactorThrottlingID:       SecondFactorThrottlingID(m.SecondFactorThrottlingID),
		AppUserID:                      AppUserID(m.AppUserID),
		FailedAttemptsSinceLastSuccess: m.FailedAttemptsSinceLastSuccess,
		TimeoutUntil:                   nullable.FromNullTime(m.TimeoutUntil),
	}
}

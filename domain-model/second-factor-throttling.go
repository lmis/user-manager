package domain_model

import (
	"time"
	"user-manager/util/nullable"

	"github.com/go-jet/jet/v2/postgres"
)

type SecondFactorThrottlingID int64

type SecondFactorThrottling struct {
	SecondFactorThrottlingID       SecondFactorThrottlingID     `json:"secondFactorThrottlingID"`
	AppUserID                      AppUserID                    `json:"appUserID"`
	FailedAttemptsSinceLastSuccess int32                        `json:"failedAttemptsSinceLastSuccess"`
	TimeoutUntil                   nullable.Nullable[time.Time] `json:"timeoutUntil,omitempty"`
}

func (a SecondFactorThrottlingID) ToIntegerExpression() postgres.IntegerExpression {
	return postgres.Int64(int64(a))
}

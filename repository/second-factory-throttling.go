package repository

import (
	"database/sql"
	"time"
	"user-manager/db"
	"user-manager/db/generated/models/postgres/public/model"
	. "user-manager/db/generated/models/postgres/public/table"
	domain_model "user-manager/domain-model"
	"user-manager/util/errors"

	. "github.com/go-jet/jet/v2/postgres"
)

type SecondFactorThrottlingRepository struct {
	tx *sql.Tx
}

func ProvideSecondFactorThrottlingRepository(tx *sql.Tx) *SecondFactorThrottlingRepository {
	return &SecondFactorThrottlingRepository{tx}
}

func (r *SecondFactorThrottlingRepository) GetForUser(userID domain_model.AppUserID) (domain_model.SecondFactorThrottling, error) {
	m, err := db.FetchMaybe[model.SecondFactorThrottling](
		SELECT(
			SecondFactorThrottling.SecondFactorThrottlingID,
			SecondFactorThrottling.AppUserID,
			SecondFactorThrottling.FailedAttemptsSinceLastSuccess,
			SecondFactorThrottling.TimeoutUntil,
		).
			FROM(SecondFactorThrottling).
			WHERE(SecondFactorThrottling.AppUserID.EQ(userID.ToIntegerExpression())).
			QueryContext,
		r.tx)
	if err != nil {
		return domain_model.SecondFactorThrottling{}, errors.Wrap("error loading throttling", err)
	}
	if m == nil {
		return domain_model.SecondFactorThrottling{}, nil
	}

	throttling := domain_model.SecondFactorThrottling{
		SecondFactorThrottlingID:       domain_model.SecondFactorThrottlingID(m.SecondFactorThrottlingID),
		AppUserID:                      domain_model.AppUserID(m.AppUserID),
		FailedAttemptsSinceLastSuccess: m.FailedAttemptsSinceLastSuccess,
	}
	if m.TimeoutUntil != nil {
		throttling.TimeoutUntil = *m.TimeoutUntil
	}
	return throttling, nil
}

func (r *SecondFactorThrottlingRepository) Update(throttlingID domain_model.SecondFactorThrottlingID, failedAttemptsSinceLastSuccess int32, maybeTimeoutUtnil *time.Time) error {
	return db.ExecSingleMutation(
		SecondFactorThrottling.UPDATE(SecondFactorThrottling.FailedAttemptsSinceLastSuccess, SecondFactorThrottling.TimeoutUntil, SecondFactorThrottling.UpdatedAt).
			SET(failedAttemptsSinceLastSuccess, maybeTimeoutUtnil, time.Now()).
			WHERE(SecondFactorThrottling.SecondFactorThrottlingID.EQ(throttlingID.ToIntegerExpression())).
			ExecContext,
		r.tx)
}

func (r *SecondFactorThrottlingRepository) Insert(userID domain_model.AppUserID, failedAttemptsSinceLastSuccess int) error {
	return db.ExecSingleMutation(
		SecondFactorThrottling.INSERT(SecondFactorThrottling.AppUserID, SecondFactorThrottling.FailedAttemptsSinceLastSuccess).
			VALUES(userID.ToIntegerExpression(), failedAttemptsSinceLastSuccess).
			ExecContext,
		r.tx)
}

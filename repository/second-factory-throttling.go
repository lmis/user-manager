package repository

import (
	"database/sql"
	"time"
	"user-manager/db"
	"user-manager/db/generated/models/postgres/public/model"
	. "user-manager/db/generated/models/postgres/public/table"
	dm "user-manager/domain-model"
	"user-manager/util/errors"

	. "github.com/go-jet/jet/v2/postgres"
)

type SecondFactorThrottlingRepository struct {
	tx *sql.Tx
}

func ProvideSecondFactorThrottlingRepository(tx *sql.Tx) *SecondFactorThrottlingRepository {
	return &SecondFactorThrottlingRepository{tx}
}

func (r *SecondFactorThrottlingRepository) GetForUser(userID dm.AppUserID) (dm.SecondFactorThrottling, error) {
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
		return dm.SecondFactorThrottling{}, errors.Wrap("error loading throttling", err)
	}
	if m == nil {
		return dm.SecondFactorThrottling{}, nil
	}

	throttling := dm.SecondFactorThrottling{
		SecondFactorThrottlingID:       dm.SecondFactorThrottlingID(m.SecondFactorThrottlingID),
		AppUserID:                      dm.AppUserID(m.AppUserID),
		FailedAttemptsSinceLastSuccess: m.FailedAttemptsSinceLastSuccess,
	}
	if m.TimeoutUntil != nil {
		throttling.TimeoutUntil = *m.TimeoutUntil
	}
	return throttling, nil
}

func (r *SecondFactorThrottlingRepository) Update(throttlingID dm.SecondFactorThrottlingID, failedAttemptsSinceLastSuccess int32, maybeTimeoutUntil *time.Time) error {
	return db.ExecSingleMutation(
		SecondFactorThrottling.UPDATE(SecondFactorThrottling.FailedAttemptsSinceLastSuccess, SecondFactorThrottling.TimeoutUntil, SecondFactorThrottling.UpdatedAt).
			SET(failedAttemptsSinceLastSuccess, maybeTimeoutUntil, time.Now()).
			WHERE(SecondFactorThrottling.SecondFactorThrottlingID.EQ(throttlingID.ToIntegerExpression())).
			ExecContext,
		r.tx)
}

func (r *SecondFactorThrottlingRepository) Insert(userID dm.AppUserID, failedAttemptsSinceLastSuccess int) error {
	return db.ExecSingleMutation(
		SecondFactorThrottling.INSERT(SecondFactorThrottling.AppUserID, SecondFactorThrottling.FailedAttemptsSinceLastSuccess).
			VALUES(userID.ToIntegerExpression(), failedAttemptsSinceLastSuccess).
			ExecContext,
		r.tx)
}

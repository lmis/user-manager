package repository

import (
	"database/sql"
	"time"
	"user-manager/db"
	"user-manager/db/generated/models/postgres/public/model"
	. "user-manager/db/generated/models/postgres/public/table"
	domain_model "user-manager/domain-model"
	"user-manager/util"
	"user-manager/util/nullable"

	. "github.com/go-jet/jet/v2/postgres"
)

type SecondFactorThrottlingRepository struct {
	tx *sql.Tx
}

func ProvideSecondFactorThrottlingRepository(tx *sql.Tx) *SecondFactorThrottlingRepository {
	return &SecondFactorThrottlingRepository{tx}
}

func (r *SecondFactorThrottlingRepository) GetForUser(userId domain_model.AppUserID) (nullable.Nullable[*domain_model.SecondFactorThrottling], error) {
	throttling, err := db.Fetch(
		SELECT(
			SecondFactorThrottling.SecondFactorThrottlingID,
			SecondFactorThrottling.AppUserID,
			SecondFactorThrottling.FailedAttemptsSinceLastSuccess,
			SecondFactorThrottling.TimeoutUntil,
		).
			FROM(SecondFactorThrottling).
			WHERE(SecondFactorThrottling.AppUserID.EQ(userId.ToIntegerExpression())).
			QueryContext,
		func(m *model.SecondFactorThrottling) *domain_model.SecondFactorThrottling {
			return &domain_model.SecondFactorThrottling{
				SecondFactorThrottlingID:       domain_model.SecondFactorThrottlingID(m.SecondFactorThrottlingID),
				AppUserID:                      domain_model.AppUserID(m.AppUserID),
				FailedAttemptsSinceLastSuccess: m.FailedAttemptsSinceLastSuccess,
				TimeoutUntil:                   nullable.FromPointer(m.TimeoutUntil),
			}
		},
		r.tx)
	if err != nil {
		return nullable.Empty[*domain_model.SecondFactorThrottling](), util.Wrap("error loading throttling", err)
	}
	return throttling, nil
}

func (r *SecondFactorThrottlingRepository) Update(throttlingId domain_model.SecondFactorThrottlingID, failedAttemptsSinceLastSuccess int32, timeoutUntil nullable.Nullable[time.Time]) error {
	return db.ExecSingleMutation(
		SecondFactorThrottling.UPDATE(SecondFactorThrottling.FailedAttemptsSinceLastSuccess, SecondFactorThrottling.TimeoutUntil, SecondFactorThrottling.UpdatedAt).
			SET(failedAttemptsSinceLastSuccess, timeoutUntil.ToPointer(), time.Now()).
			WHERE(SecondFactorThrottling.SecondFactorThrottlingID.EQ(throttlingId.ToIntegerExpression())).
			ExecContext,
		r.tx)
}

func (r *SecondFactorThrottlingRepository) Insert(userId domain_model.AppUserID, failedAttemptsSinceLastSuccess int) error {
	return db.ExecSingleMutation(
		SecondFactorThrottling.INSERT(SecondFactorThrottling.AppUserID, SecondFactorThrottling.FailedAttemptsSinceLastSuccess).
			VALUES(userId.ToIntegerExpression(), failedAttemptsSinceLastSuccess).
			ExecContext,
		r.tx)
}

package repository

import (
	"context"
	"database/sql"
	"time"
	"user-manager/db"
	"user-manager/db/generated/models"
	domain_model "user-manager/domain-model"
	"user-manager/util"
	"user-manager/util/nullable"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type SecondFactorThrottlingRepository struct {
	tx *sql.Tx
}

func ProvideSecondFactorThrottlingRepository(tx *sql.Tx) *SecondFactorThrottlingRepository {
	return &SecondFactorThrottlingRepository{tx}
}

func (r *SecondFactorThrottlingRepository) GetForUser(userId domain_model.AppUserID) (nullable.Nullable[*domain_model.SecondFactorThrottling], error) {
	throttling, err := db.Fetch(func(ctx context.Context) (*models.SecondFactorThrottling, error) {
		return models.SecondFactorThrottlings(
			models.SecondFactorThrottlingWhere.AppUserID.EQ(int64(userId)),
		).One(ctx, r.tx)
	})
	if err != nil {
		return nullable.Empty[*domain_model.SecondFactorThrottling](), util.Wrap("error loading throttling", err)
	}
	if throttling.IsEmpty() {
		return nullable.Empty[*domain_model.SecondFactorThrottling](), nil
	}
	return nullable.NeverNil(domain_model.FromSecondFactorThrottlingModel(throttling.OrPanic())), nil
}
func (r *SecondFactorThrottlingRepository) Update(throttlingId domain_model.SecondFactorThrottlingID, failedAttemptsSinceLastSuccess int, timeoutUntil nullable.Nullable[time.Time]) error {
	return db.ExecSingleMutation(func(ctx context.Context) (int64, error) {
		throttling := models.SecondFactorThrottling{SecondFactorThrottlingID: int64(throttlingId), FailedAttemptsSinceLastSuccess: failedAttemptsSinceLastSuccess}
		if timeoutUntil.IsPresent {
			throttling.TimeoutUntil = null.TimeFrom(timeoutUntil.OrPanic())
		}
		return throttling.Update(ctx, r.tx, boil.Whitelist(models.SecondFactorThrottlingColumns.FailedAttemptsSinceLastSuccess, models.SecondFactorThrottlingColumns.TimeoutUntil))
	})
}

func (r *SecondFactorThrottlingRepository) Insert(userId domain_model.AppUserID, failedAttemptsSinceLastSuccess int) error {
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()
	throttling := models.SecondFactorThrottling{AppUserID: int64(userId), FailedAttemptsSinceLastSuccess: failedAttemptsSinceLastSuccess}
	return throttling.Insert(ctx, r.tx, boil.Infer())
}

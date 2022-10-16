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
	throttling, err := db.Fetch(func(ctx context.Context) (*models.TwoFactorThrottling, error) {
		return models.TwoFactorThrottlings(
			models.TwoFactorThrottlingWhere.AppUserID.EQ(int64(userId)),
		).One(ctx, r.tx)
	})
	if err != nil {
		return nullable.Empty[*domain_model.SecondFactorThrottling](), util.Wrap("error loading throttling", err)
	}
	if throttling.IsEmpty() {
		return nullable.Empty[*domain_model.SecondFactorThrottling](), nil
	}
	return nullable.NeverNil(domain_model.FromSecondFactorThrottlingModel(throttling.Val)), nil
}
func (r *SecondFactorThrottlingRepository) Update(throttlingId domain_model.TwoFactorThrottlingID, failedAttemptsSinceLastSuccess int, timeoutUntil nullable.Nullable[time.Time]) error {
	return db.ExecSingleMutation(func(ctx context.Context) (int64, error) {
		throttling := models.TwoFactorThrottling{TwoFactorThrottlingID: int64(throttlingId), FailedAttemptsSinceLastSuccess: failedAttemptsSinceLastSuccess}
		if timeoutUntil.IsPresent {
			throttling.TimeoutUntil = null.TimeFrom(timeoutUntil.Val)
		}
		return throttling.Update(ctx, r.tx, boil.Whitelist(models.TwoFactorThrottlingColumns.FailedAttemptsSinceLastSuccess, models.TwoFactorThrottlingColumns.TimeoutUntil))
	})
}

func (r *SecondFactorThrottlingRepository) Insert(userId domain_model.AppUserID, failedAttemptsSinceLastSuccess int) error {
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()
	throttling := models.TwoFactorThrottling{AppUserID: int64(userId), FailedAttemptsSinceLastSuccess: failedAttemptsSinceLastSuccess}
	return throttling.Insert(ctx, r.tx, boil.Infer())
}

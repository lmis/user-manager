package user

import (
	"context"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db"
	"user-manager/db/generated/models"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

func UpdateUser(requestContext *ginext.RequestContext, user *models.AppUser) error {
	return db.ExecSingleMutation(func(ctx context.Context) (int64, error) { return user.Update(ctx, requestContext.Tx, boil.Infer()) })
}

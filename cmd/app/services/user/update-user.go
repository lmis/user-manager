package userservice

import (
	"fmt"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/util"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

func UpdateUser(requestContext *ginext.RequestContext, user *models.AppUser) error {
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()
	rows, err := user.Update(ctx, requestContext.Tx, boil.Infer())
	if err != nil {
		return util.Wrap("issue updating user in db", err)
	}
	if rows != 1 {
		return util.Wrap(fmt.Sprintf("wrong number of rows affected: %d", rows), err)
	}

	return nil
}

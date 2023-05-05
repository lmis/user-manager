package middleware

import (
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db"
	"user-manager/util/errors"

	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterDatabaseMiddleware(group *gin.RouterGroup, database *sql.DB) error {
	if database == nil {
		return errors.Error("Invalid middleware setup: db is nil")
	}
	group.Use(func(ctx *gin.Context) {
		r := ginext.GetRequestContext(ctx)
		log := r.Log

		dbCtx, cancelTimeout := db.DefaultQueryContext(ctx)
		defer cancelTimeout()

		log.Info("BEGIN Transaction")
		tx, err := database.BeginTx(dbCtx, nil)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, errors.Wrap("begin transaction failed", err))
			return
		}

		defer func() {
			if !ctx.IsAborted() {
				log.Info("COMMIT")
				if err := tx.Commit(); err != nil {
					_ = ctx.AbortWithError(http.StatusInternalServerError, errors.Wrap("commit failed", err))
					return
				}
			} else {
				log.Info("ROLLBACK")
				if err = tx.Rollback(); err != nil {
					// If rollback doesn't work, log and forget
					_ = ctx.AbortWithError(http.StatusInternalServerError, errors.Wrap("rollback failed", err))
					return
				}
			}
		}()

		r.Tx = tx

		ctx.Next()
	})
	return nil
}

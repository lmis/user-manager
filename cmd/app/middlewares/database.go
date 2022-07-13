package middlewares

import (
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db"
	"user-manager/util"

	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DatabaseMiddleware(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestCtx := ginext.GetRequestContext(c)
		log := requestCtx.Log
		ctx, cancelTimeout := db.DefaultQueryContext()
		defer cancelTimeout()

		log.Info("BEGIN Transaction")
		tx, err := database.BeginTx(ctx, nil)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, util.Wrap("begin transaction failed", err))
			return
		}

		defer func() {
			if !c.IsAborted() {
				log.Info("COMMIT")
				if err := tx.Commit(); err != nil {
					c.AbortWithError(http.StatusInternalServerError, util.Wrap("commit failed", err))
				}
			} else {
				log.Info("ROLLBACK")
				if err = tx.Rollback(); err != nil {
					// If rollback doesn't work, log and forget
					c.AbortWithError(http.StatusInternalServerError, util.Wrap("rollback failed", err))
				}
			}
		}()
		requestCtx.Tx = tx

		c.Next()
	}
}

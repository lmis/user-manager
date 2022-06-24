package middlewares

import (
	"user-manager/db"
	ginext "user-manager/gin-extensions"
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
			c.AbortWithError(http.StatusInternalServerError, util.Wrap("DatabaseMiddleware", "begin transaction failed", err))
			return
		}

		defer func() {
			log.Info("ROLLBACK")
			err = tx.Rollback()
			if err != nil {
				// If rollback doesn't work, log and forget
				c.Error(util.Wrap("DatabaseMiddleware", "rollback failed", err))
			}
		}()
		requestCtx.Tx = tx

		c.Next()

		if !c.IsAborted() {
			log.Info("COMMIT")
			if err := tx.Commit(); err != nil {
				c.AbortWithError(http.StatusInternalServerError, util.Wrap("DatabaseMiddlware", "commit failed", err))
			}
		}
	}
}

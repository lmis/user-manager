package middlewares

import (
	"user-manager/db"
	ginext "user-manager/gin-extensions"

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
			log.Err(err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		defer func() {
			log.Info("ROLLBACK")
			tx.Rollback()
		}()
		requestCtx.Tx = tx

		c.Next()

		if !c.IsAborted() {
			log.Info("COMMIT")
			if err := tx.Commit(); err != nil {
				ginext.LogAndAbortWithError(c, http.StatusInternalServerError, err)
			}
		}
	}
}

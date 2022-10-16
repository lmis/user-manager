package middleware

import (
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db"
	domain_model "user-manager/domain-model"
	"user-manager/util"

	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DatabaseMiddleware struct {
	c        *gin.Context
	database *sql.DB
	log      domain_model.Log
}

func ProvideDatabaseMiddleware(c *gin.Context, database *sql.DB, log domain_model.Log) *DatabaseMiddleware {
	return &DatabaseMiddleware{c, database, log}
}

func RegisterDatabaseMiddleware(group *gin.RouterGroup) {
	group.Use(ginext.WrapMiddleware(InitializeDatabaseMiddleware))
}
func (m *DatabaseMiddleware) Handle() {
	c := m.c
	log := m.log

	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()

	log.Info("BEGIN Transaction")
	tx, err := m.database.BeginTx(ctx, nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("begin transaction failed", err))
		return
	}

	defer func() {
		if !c.IsAborted() {
			log.Info("COMMIT")
			if err := tx.Commit(); err != nil {
				c.AbortWithError(http.StatusInternalServerError, util.Wrap("commit failed", err))
				return
			}
		} else {
			log.Info("ROLLBACK")
			if err = tx.Rollback(); err != nil {
				// If rollback doesn't work, log and forget
				c.AbortWithError(http.StatusInternalServerError, util.Wrap("rollback failed", err))
				return
			}
		}
	}()

	ginext.GetRequestContext(c).Tx = tx

	c.Next()
}

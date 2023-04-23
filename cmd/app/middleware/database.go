package middleware

import (
	"context"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db"
	dm "user-manager/domain-model"
	"user-manager/util/errors"

	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DatabaseMiddleware struct {
	c        *gin.Context
	ctx      context.Context
	database *sql.DB
	log      dm.Log
}

func ProvideDatabaseMiddleware(c *gin.Context, ctx context.Context, database *sql.DB, log dm.Log) *DatabaseMiddleware {
	return &DatabaseMiddleware{c, ctx, database, log}
}

func RegisterDatabaseMiddleware(group *gin.RouterGroup) {
	group.Use(func(ctx *gin.Context) { InitializeDatabaseMiddleware(ctx).Handle() })
}
func (m *DatabaseMiddleware) Handle() {
	c := m.c
	log := m.log
	database := m.database

	ctx, cancelTimeout := db.DefaultQueryContext(m.ctx)
	defer cancelTimeout()

	log.Info("BEGIN Transaction")
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, errors.Wrap("begin transaction failed", err))
		return
	}

	defer func() {
		if !c.IsAborted() {
			log.Info("COMMIT")
			if err := tx.Commit(); err != nil {
				_ = c.AbortWithError(http.StatusInternalServerError, errors.Wrap("commit failed", err))
				return
			}
		} else {
			log.Info("ROLLBACK")
			if err = tx.Rollback(); err != nil {
				// If rollback doesn't work, log and forget
				_ = c.AbortWithError(http.StatusInternalServerError, errors.Wrap("rollback failed", err))
				return
			}
		}
	}()

	ginext.GetRequestContext(c).Tx = tx

	c.Next()
}

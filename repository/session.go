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

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type SessionRepository struct {
	tx *sql.Tx
}

func ProvideSessionRepository(tx *sql.Tx) *SessionRepository {
	return &SessionRepository{tx}
}

func (r *SessionRepository) InsertSession(sessionId string, sessionType domain_model.UserSessionType, appUserId domain_model.AppUserID, duration time.Duration) error {
	session := models.UserSession{
		UserSessionID:   sessionId,
		AppUserID:       int64(appUserId),
		TimeoutAt:       time.Now().Add(duration),
		UserSessionType: models.UserSessionType(sessionType),
	}
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()
	if err := session.Insert(ctx, r.tx, boil.Infer()); err != nil {
		return util.Wrap("cannot insert session", err)
	}
	return nil
}

func (r *SessionRepository) UpdateSessionTimeout(sessionId domain_model.UserSessionID, timeout time.Time) error {
	if err := db.ExecSingleMutation(func(ctx context.Context) (int64, error) {
		return (&models.UserSession{UserSessionID: string(sessionId), TimeoutAt: timeout}).Update(ctx, r.tx, boil.Whitelist(models.UserSessionColumns.TimeoutAt))
	}); err != nil {
		return util.Wrap("error updating session timeout", err)
	}
	return nil
}

func (r *SessionRepository) GetSessionAndUser(sessionId string, sessionType domain_model.UserSessionType) (nullable.Nullable[*domain_model.UserSession], error) {
	session, err := db.Fetch(func(ctx context.Context) (*models.UserSession, error) {
		return models.UserSessions(models.UserSessionWhere.UserSessionID.EQ(sessionId),
			models.UserSessionWhere.TimeoutAt.GT(time.Now()),
			models.UserSessionWhere.UserSessionType.EQ(models.UserSessionType(sessionType)),
			qm.Load(models.UserSessionRels.AppUser)).
			One(ctx, r.tx)
	})

	if err != nil {
		return nullable.Empty[*domain_model.UserSession](), util.Wrap("error finding session with user", err)
	}
	if session.IsEmpty() {
		return nullable.Empty[*domain_model.UserSession](), nil
	}

	roles, err := db.Fetch(func(ctx context.Context) (models.AppUserRoleSlice, error) {
		return models.AppUserRoles(models.AppUserRoleWhere.AppUserID.EQ(session.Val.AppUserID)).
			All(ctx, r.tx)
	})

	if err != nil {
		return nullable.Empty[*domain_model.UserSession](), util.Wrap("error fetching user role", err)
	}

	if roles.IsEmpty() {
		return nullable.Empty[*domain_model.UserSession](), nil
	}

	return nullable.NeverNil(domain_model.FromUserSessionAppUserAndUserRolesModel(session.Val, session.Val.R.AppUser, roles.Val)), nil
}

package repository

import (
	"context"
	"database/sql"
	"time"
	"user-manager/db"
	"user-manager/db/generated/models/postgres/public/model"
	. "user-manager/db/generated/models/postgres/public/table"
	dm "user-manager/domain-model"
	"user-manager/util/errors"

	. "github.com/go-jet/jet/v2/postgres"
)

type SessionRepository struct {
	ctx context.Context
	tx  *sql.Tx
}

func ProvideSessionRepository(ctx context.Context, tx *sql.Tx) *SessionRepository {
	return &SessionRepository{ctx, tx}
}

func (r *SessionRepository) InsertSession(sessionID string, sessionType dm.UserSessionType, appUserID dm.AppUserID, duration time.Duration) error {
	return db.ExecSingleMutation(
		r.ctx,
		UserSession.INSERT(UserSession.UserSessionID, UserSession.AppUserID, UserSession.TimeoutAt, UserSession.UserSessionType).
			VALUES(sessionID, appUserID.ToIntegerExpression(), time.Now().Add(duration), model.UserSessionType(sessionType)).
			ExecContext,
		r.tx)
}

func (r *SessionRepository) UpdateSessionTimeout(sessionID dm.UserSessionID, timeout time.Time) error {
	return db.ExecSingleMutation(
		r.ctx,
		UserSession.UPDATE(UserSession.TimeoutAt, UserSession.UpdatedAt).
			SET(timeout, time.Now()).
			WHERE(UserSession.UserSessionID.EQ(sessionID.ToStringExpression())).
			ExecContext,
		r.tx)
}

func (r *SessionRepository) Delete(sessionID dm.UserSessionID) error {
	return db.ExecSingleMutation(
		r.ctx,
		UserSession.DELETE().
			WHERE(UserSession.UserSessionID.EQ(sessionID.ToStringExpression())).
			ExecContext,
		r.tx)
}

func (r *SessionRepository) GetSessionAndUser(sessionID dm.UserSessionID, sessionType dm.UserSessionType) (dm.UserSession, error) {
	m, err := db.FetchMaybe[struct {
		model.UserSession
		model.AppUser
		Roles []model.AppUserRole
	}](
		r.ctx,
		SELECT(
			UserSession.UserSessionID,
			UserSession.UserSessionType,
			AppUser.AppUserID,
			AppUser.Language,
			AppUser.UserName,
			AppUser.PasswordHash,
			AppUser.Email,
			AppUser.EmailVerified,
			AppUser.EmailVerificationToken,
			AppUser.NextEmail,
			AppUser.PasswordResetToken,
			AppUser.PasswordResetTokenValidUntil,
			AppUser.SecondFactorToken,
			AppUser.TemporarySecondFactorToken,
			AppUserRole.Role,
		).
			FROM(UserSession.
				INNER_JOIN(AppUser, AppUser.AppUserID.EQ(UserSession.AppUserID)).
				INNER_JOIN(AppUserRole, AppUserRole.AppUserID.EQ(AppUser.AppUserID)),
			).
			WHERE(
				UserSession.UserSessionID.EQ(sessionID.ToStringExpression()).
					AND(UserSession.TimeoutAt.GT(TimestampzT(time.Now()))).
					AND(UserSession.UserSessionType.EQ(sessionType.ToStringExpression())),
			).
			QueryContext,
		r.tx)
	if err != nil {
		return dm.UserSession{}, errors.Wrap("error loading user session", err)
	}

	if m == nil {
		return dm.UserSession{}, nil
	}

	userSession := dm.UserSession{
		UserSessionID: dm.UserSessionID(m.UserSessionID),
		User: &dm.AppUser{
			AppUserID:                  dm.AppUserID(m.AppUser.AppUserID),
			Language:                   dm.UserLanguage(m.Language),
			UserName:                   m.UserName,
			PasswordHash:               m.PasswordHash,
			Email:                      m.Email,
			EmailVerified:              m.EmailVerified,
			EmailVerificationToken:     m.EmailVerificationToken,
			NextEmail:                  m.NextEmail,
			PasswordResetToken:         m.PasswordResetToken,
			SecondFactorToken:          m.SecondFactorToken,
			TemporarySecondFactorToken: m.TemporarySecondFactorToken,
			UserRoles:                  make([]dm.UserRole, len(m.Roles)),
		},
		UserSessionType: dm.UserSessionType(m.UserSessionType),
	}
	if m.PasswordResetTokenValidUntil != nil {
		userSession.User.PasswordResetTokenValidUntil = *m.PasswordResetTokenValidUntil
	}
	for i, role := range m.Roles {
		userSession.User.UserRoles[i] = dm.UserRole(role.Role)
	}
	return userSession, nil
}

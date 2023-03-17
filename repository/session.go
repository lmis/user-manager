package repository

import (
	"database/sql"
	"time"
	"user-manager/db"
	"user-manager/db/generated/models/postgres/public/model"
	. "user-manager/db/generated/models/postgres/public/table"
	domain_model "user-manager/domain-model"
	"user-manager/util/errors"

	. "github.com/go-jet/jet/v2/postgres"
)

type SessionRepository struct {
	tx *sql.Tx
}

func ProvideSessionRepository(tx *sql.Tx) *SessionRepository {
	return &SessionRepository{tx}
}

func (r *SessionRepository) InsertSession(sessionID string, sessionType domain_model.UserSessionType, appUserID domain_model.AppUserID, duration time.Duration) error {
	return db.ExecSingleMutation(
		UserSession.INSERT(UserSession.UserSessionID, UserSession.AppUserID, UserSession.TimeoutAt, UserSession.UserSessionType).
			VALUES(sessionID, appUserID.ToIntegerExpression(), time.Now().Add(duration), model.UserSessionType(sessionType)).
			ExecContext,
		r.tx)
}

func (r *SessionRepository) UpdateSessionTimeout(sessionID domain_model.UserSessionID, timeout time.Time) error {
	return db.ExecSingleMutation(
		UserSession.UPDATE(UserSession.TimeoutAt, UserSession.UpdatedAt).
			SET(timeout, time.Now()).
			WHERE(UserSession.UserSessionID.EQ(sessionID.ToStringExpression())).
			ExecContext,
		r.tx)
}

func (r *SessionRepository) Delete(sessionID domain_model.UserSessionID) error {
	return db.ExecSingleMutation(
		UserSession.DELETE().
			WHERE(UserSession.UserSessionID.EQ(sessionID.ToStringExpression())).
			ExecContext,
		r.tx)
}

func (r *SessionRepository) GetSessionAndUser(sessionID domain_model.UserSessionID, sessionType domain_model.UserSessionType) (domain_model.UserSession, error) {
	m, err := db.FetchMaybe[struct {
		model.UserSession
		model.AppUser
		Roles []model.AppUserRole
	}](
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
		return domain_model.UserSession{}, errors.Wrap("error loading user session", err)
	}

	if m == nil {
		return domain_model.UserSession{}, nil
	}

	userSession := domain_model.UserSession{
		UserSessionID: domain_model.UserSessionID(m.UserSessionID),
		User: &domain_model.AppUser{
			AppUserID:                  domain_model.AppUserID(m.AppUser.AppUserID),
			Language:                   domain_model.UserLanguage(m.Language),
			UserName:                   m.UserName,
			PasswordHash:               m.PasswordHash,
			Email:                      m.Email,
			EmailVerified:              m.EmailVerified,
			EmailVerificationToken:     m.EmailVerificationToken,
			NextEmail:                  m.NextEmail,
			PasswordResetToken:         m.PasswordResetToken,
			SecondFactorToken:          m.SecondFactorToken,
			TemporarySecondFactorToken: m.TemporarySecondFactorToken,
			UserRoles:                  make([]domain_model.UserRole, len(m.Roles)),
		},
		UserSessionType: domain_model.UserSessionType(m.UserSessionType),
	}
	if m.PasswordResetTokenValidUntil != nil {
		userSession.User.PasswordResetTokenValidUntil = *m.PasswordResetTokenValidUntil
	}
	for i, role := range m.Roles {
		userSession.User.UserRoles[i] = domain_model.UserRole(role.Role)
	}
	return userSession, nil
}

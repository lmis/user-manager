package repository

import (
	"database/sql"
	"time"
	"user-manager/db"
	"user-manager/db/generated/models/postgres/public/model"
	. "user-manager/db/generated/models/postgres/public/table"
	domain_model "user-manager/domain-model"
	"user-manager/util/nullable"
	"user-manager/util/slices"

	. "github.com/go-jet/jet/v2/postgres"
)

type SessionRepository struct {
	tx *sql.Tx
}

func ProvideSessionRepository(tx *sql.Tx) *SessionRepository {
	return &SessionRepository{tx}
}

func (r *SessionRepository) InsertSession(sessionId string, sessionType domain_model.UserSessionType, appUserId domain_model.AppUserID, duration time.Duration) error {
	return db.ExecSingleMutation(
		UserSession.INSERT(UserSession.UserSessionID, UserSession.AppUserID, UserSession.TimeoutAt, UserSession.UserSessionType).
			VALUES(sessionId, appUserId.ToIntegerExpression(), time.Now().Add(duration), model.UserSessionType(sessionType)).
			ExecContext,
		r.tx)
}

func (r *SessionRepository) UpdateSessionTimeout(sessionId domain_model.UserSessionID, timeout time.Time) error {
	return db.ExecSingleMutation(
		UserSession.UPDATE(UserSession.TimeoutAt, UserSession.UpdatedAt).
			SET(timeout, time.Now()).
			WHERE(UserSession.UserSessionID.EQ(sessionId.ToStringExpression())).
			ExecContext,
		r.tx)
}

func (r *SessionRepository) Delete(sessionId domain_model.UserSessionID) error {
	return db.ExecSingleMutation(
		UserSession.DELETE().
			WHERE(UserSession.UserSessionID.EQ(sessionId.ToStringExpression())).
			ExecContext,
		r.tx)
}

func (r *SessionRepository) GetSessionAndUser(sessionId domain_model.UserSessionID, sessionType domain_model.UserSessionType) (nullable.Nullable[domain_model.UserSession], error) {
	return db.Fetch(
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
				UserSession.UserSessionID.EQ(sessionId.ToStringExpression()).
					AND(UserSession.TimeoutAt.GT(TimestampzT(time.Now()))).
					AND(UserSession.UserSessionType.EQ(sessionType.ToStringExpression())),
			).
			QueryContext,
		func(m *struct {
			model.UserSession
			model.AppUser
			Roles []model.AppUserRole
		}) domain_model.UserSession {
			return domain_model.UserSession{
				UserSessionID: domain_model.UserSessionID(m.UserSessionID),
				User: &domain_model.AppUser{
					AppUserID:                    domain_model.AppUserID(m.AppUser.AppUserID),
					Language:                     domain_model.UserLanguage(m.Language),
					UserName:                     m.UserName,
					PasswordHash:                 m.PasswordHash,
					Email:                        m.Email,
					EmailVerified:                m.EmailVerified,
					EmailVerificationToken:       nullable.FromPointer(m.EmailVerificationToken),
					NextEmail:                    nullable.FromPointer(m.NextEmail),
					PasswordResetToken:           nullable.FromPointer(m.PasswordResetToken),
					PasswordResetTokenValidUntil: nullable.FromPointer(m.PasswordResetTokenValidUntil),
					SecondFactorToken:            nullable.FromPointer(m.SecondFactorToken),
					TemporarySecondFactorToken:   nullable.FromPointer(m.TemporarySecondFactorToken),
					UserRoles:                    slices.Map(m.Roles, func(r model.AppUserRole) domain_model.UserRole { return domain_model.UserRole(r.Role) }),
				},
				UserSessionType: domain_model.UserSessionType(m.UserSessionType),
			}
		},
		r.tx)
}

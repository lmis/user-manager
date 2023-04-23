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

type UserRepository struct {
	ctx context.Context
	tx  *sql.Tx
}

func ProvideUserRepository(ctx context.Context, tx *sql.Tx) *UserRepository {
	return &UserRepository{ctx, tx}
}

func (r *UserRepository) GetUserForEmail(email string) (dm.AppUser, error) {
	m, err := db.FetchMaybe[struct {
		model.AppUser
		Roles []model.AppUserRole
	}](
		r.ctx,
		SELECT(
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
			FROM(AppUser.
				INNER_JOIN(AppUserRole, AppUserRole.AppUserID.EQ(AppUser.AppUserID)),
			).
			WHERE(AppUser.Email.EQ(String(email))).
			QueryContext,
		r.tx)
	if err != nil {
		return dm.AppUser{}, errors.Wrap("error loading user", err)
	}

	if m == nil {
		return dm.AppUser{}, nil
	}
	user := dm.AppUser{
		AppUserID:                  dm.AppUserID(m.AppUserID),
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
	}
	if m.PasswordResetTokenValidUntil != nil {
		user.PasswordResetTokenValidUntil = *m.PasswordResetTokenValidUntil
	}

	for i, role := range m.Roles {
		user.UserRoles[i] = dm.UserRole(role.Role)
	}
	return user, nil
}

func (r *UserRepository) UpdateUserEmailVerificationToken(appUserID dm.AppUserID, token string) error {
	return db.ExecSingleMutation(
		r.ctx,
		AppUser.UPDATE(AppUser.EmailVerificationToken, AppUser.UpdatedAt).
			SET(token, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserID.ToIntegerExpression())).
			ExecContext,
		r.tx)
}

func (r *UserRepository) SetEmailToVerified(appUserID dm.AppUserID) error {
	return db.ExecSingleMutation(
		r.ctx,
		AppUser.UPDATE(AppUser.EmailVerificationToken, AppUser.EmailVerified, AppUser.UpdatedAt).
			SET("", true, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserID.ToIntegerExpression())).
			ExecContext,
		r.tx)
}

func (r *UserRepository) SetNextEmail(appUserID dm.AppUserID, nextEmail string, verificationToken string) error {
	return db.ExecSingleMutation(
		r.ctx,
		AppUser.UPDATE(AppUser.EmailVerificationToken, AppUser.NextEmail, AppUser.UpdatedAt).
			SET(verificationToken, nextEmail, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserID.ToIntegerExpression())).
			ExecContext,
		r.tx)
}

func (r *UserRepository) SetEmailAndClearNextEmail(appUserID dm.AppUserID, email string) error {
	return db.ExecSingleMutation(
		r.ctx,
		AppUser.UPDATE(AppUser.EmailVerificationToken, AppUser.NextEmail, AppUser.Email, AppUser.UpdatedAt).
			SET("", "", email, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserID.ToIntegerExpression())).
			ExecContext,
		r.tx)
}

func (r *UserRepository) SetPasswordResetToken(appUserID dm.AppUserID, token string, validUntil time.Time) error {
	return db.ExecSingleMutation(
		r.ctx,
		AppUser.UPDATE(AppUser.PasswordResetToken, AppUser.PasswordResetTokenValidUntil, AppUser.UpdatedAt).
			SET(token, validUntil, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserID.ToIntegerExpression())).
			ExecContext,
		r.tx)
}

func (r *UserRepository) SetPasswordHash(appUserID dm.AppUserID, hash string) error {
	return db.ExecSingleMutation(
		r.ctx,
		AppUser.UPDATE(AppUser.PasswordHash, AppUser.PasswordResetToken, AppUser.PasswordResetTokenValidUntil, AppUser.UpdatedAt).
			SET(hash, "", nil, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserID.ToIntegerExpression())).
			ExecContext,
		r.tx)
}

func (r *UserRepository) SetLanguage(appUserID dm.AppUserID, language dm.UserLanguage) error {
	return db.ExecSingleMutation(
		r.ctx,
		AppUser.UPDATE(AppUser.Language, AppUser.UpdatedAt).
			SET(model.UserLanguage(language), time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserID.ToIntegerExpression())).
			ExecContext,
		r.tx)
}

func (r *UserRepository) Insert(userRole dm.UserRole, userName string, email string, emailVerified bool, emailVerificationToken string, passwordHash string, language dm.UserLanguage) error {
	stmt := AppUser.INSERT(
		AppUser.UserName,
		AppUser.Email,
		AppUser.EmailVerified,
		AppUser.EmailVerificationToken,
		AppUser.PasswordHash,
		AppUser.Language,
	).
		VALUES(userName, email, emailVerified, emailVerificationToken, passwordHash, model.UserLanguage(language)).
		RETURNING(AppUser.AppUserID)

	ctx, cancelTimeout := db.DefaultQueryContext(r.ctx)
	defer cancelTimeout()
	res := &model.AppUser{}
	err := stmt.QueryContext(ctx, r.tx, res)
	if err != nil {
		return errors.Wrap("cannot insert user", err)
	}

	return db.ExecSingleMutation(
		r.ctx,
		AppUserRole.INSERT(AppUserRole.Role, AppUserRole.AppUserID).
			VALUES(userRole, res.AppUserID).
			ExecContext,
		r.tx)
}

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

func GetUserForEmail(ctx context.Context, tx *sql.Tx, email string) (dm.AppUser, error) {
	m, err := db.FetchMaybe[struct {
		model.AppUser
		Roles []model.AppUserRole
	}](
		ctx,
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
		tx)
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

func UpdateUserEmailVerificationToken(ctx context.Context, tx *sql.Tx, appUserID dm.AppUserID, token string) error {
	return db.ExecSingleMutation(
		ctx,
		AppUser.UPDATE(AppUser.EmailVerificationToken, AppUser.UpdatedAt).
			SET(token, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserID.ToIntegerExpression())).
			ExecContext,
		tx)
}

func SetEmailToVerified(ctx context.Context, tx *sql.Tx, appUserID dm.AppUserID) error {
	return db.ExecSingleMutation(
		ctx,
		AppUser.UPDATE(AppUser.EmailVerificationToken, AppUser.EmailVerified, AppUser.UpdatedAt).
			SET("", true, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserID.ToIntegerExpression())).
			ExecContext,
		tx)
}

func SetNextEmail(ctx context.Context, tx *sql.Tx, appUserID dm.AppUserID, nextEmail string, verificationToken string) error {
	return db.ExecSingleMutation(
		ctx,
		AppUser.UPDATE(AppUser.EmailVerificationToken, AppUser.NextEmail, AppUser.UpdatedAt).
			SET(verificationToken, nextEmail, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserID.ToIntegerExpression())).
			ExecContext,
		tx)
}

func SetEmailAndClearNextEmail(ctx context.Context, tx *sql.Tx, appUserID dm.AppUserID, email string) error {
	return db.ExecSingleMutation(
		ctx,
		AppUser.UPDATE(AppUser.EmailVerificationToken, AppUser.NextEmail, AppUser.Email, AppUser.UpdatedAt).
			SET("", "", email, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserID.ToIntegerExpression())).
			ExecContext,
		tx)
}

func SetPasswordResetToken(ctx context.Context, tx *sql.Tx, appUserID dm.AppUserID, token string, validUntil time.Time) error {
	return db.ExecSingleMutation(
		ctx,
		AppUser.UPDATE(AppUser.PasswordResetToken, AppUser.PasswordResetTokenValidUntil, AppUser.UpdatedAt).
			SET(token, validUntil, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserID.ToIntegerExpression())).
			ExecContext,
		tx)
}

func SetPasswordHash(ctx context.Context, tx *sql.Tx, appUserID dm.AppUserID, hash string) error {
	return db.ExecSingleMutation(
		ctx,
		AppUser.UPDATE(AppUser.PasswordHash, AppUser.PasswordResetToken, AppUser.PasswordResetTokenValidUntil, AppUser.UpdatedAt).
			SET(hash, "", nil, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserID.ToIntegerExpression())).
			ExecContext,
		tx)
}

func SetLanguage(ctx context.Context, tx *sql.Tx, appUserID dm.AppUserID, language dm.UserLanguage) error {
	return db.ExecSingleMutation(
		ctx,
		AppUser.UPDATE(AppUser.Language, AppUser.UpdatedAt).
			SET(model.UserLanguage(language), time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserID.ToIntegerExpression())).
			ExecContext,
		tx)
}

func InsertUser(ctx context.Context, tx *sql.Tx, userRole dm.UserRole, userName string, email string, emailVerified bool, emailVerificationToken string, passwordHash string, language dm.UserLanguage) error {
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

	ctx, cancelTimeout := db.DefaultQueryContext(ctx)
	defer cancelTimeout()
	res := &model.AppUser{}
	err := stmt.QueryContext(ctx, tx, res)
	if err != nil {
		return errors.Wrap("cannot insert user", err)
	}

	return db.ExecSingleMutation(
		ctx,
		AppUserRole.INSERT(AppUserRole.Role, AppUserRole.AppUserID).
			VALUES(userRole, res.AppUserID).
			ExecContext,
		tx)
}

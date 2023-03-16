package repository

import (
	"database/sql"
	"time"
	"user-manager/db"
	. "user-manager/db/generated/models/postgres/public/enum"
	"user-manager/db/generated/models/postgres/public/model"
	. "user-manager/db/generated/models/postgres/public/table"
	domain_model "user-manager/domain-model"
	"user-manager/util/errors"
	"user-manager/util/nullable"

	. "github.com/go-jet/jet/v2/postgres"
)

type UserRepository struct {
	tx *sql.Tx
}

func ProvideUserRepository(tx *sql.Tx) *UserRepository {
	return &UserRepository{tx}
}

func (r *UserRepository) GetUserForEmail(email string) (nullable.Nullable[domain_model.AppUser], error) {
	maybeModel, err := db.Fetch[struct {
		model.AppUser
		Roles []model.AppUserRole
	}](
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
		return nullable.Empty[domain_model.AppUser](), errors.Wrap("error loading user", err)
	}

	if maybeModel.IsEmpty() {
		return nullable.Empty[domain_model.AppUser](), nil
	}
	m := maybeModel.OrPanic()
	user := domain_model.AppUser{
		AppUserID:                    domain_model.AppUserID(m.AppUserID),
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
		UserRoles:                    make([]domain_model.UserRole, len(m.Roles)),
	}
	for i, role := range m.Roles {
		user.UserRoles[i] = domain_model.UserRole(role.Role)
	}
	return nullable.Of(user), nil
}

func (r *UserRepository) UpdateUserEmailVerificationToken(appUserId domain_model.AppUserID, token string) error {
	return db.ExecSingleMutation(
		AppUser.UPDATE(AppUser.EmailVerificationToken, AppUser.UpdatedAt).
			SET(token, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserId.ToIntegerExpression())).
			ExecContext,
		r.tx)
}

func (r *UserRepository) SetEmailToVerified(appUserId domain_model.AppUserID) error {
	return db.ExecSingleMutation(
		AppUser.UPDATE(AppUser.EmailVerificationToken, AppUser.EmailVerified, AppUser.UpdatedAt).
			SET(nil, true, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserId.ToIntegerExpression())).
			ExecContext,
		r.tx)
}

func (r *UserRepository) SetNextEmail(appUserId domain_model.AppUserID, nextEmail string, verificationToken string) error {
	return db.ExecSingleMutation(
		AppUser.UPDATE(AppUser.EmailVerificationToken, AppUser.NextEmail, AppUser.UpdatedAt).
			SET(verificationToken, nextEmail, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserId.ToIntegerExpression())).
			ExecContext,
		r.tx)
}

func (r *UserRepository) SetEmailAndClearNextEmail(appUserId domain_model.AppUserID, email string) error {
	return db.ExecSingleMutation(
		AppUser.UPDATE(AppUser.EmailVerificationToken, AppUser.NextEmail, AppUser.Email, AppUser.UpdatedAt).
			SET(nil, nil, email, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserId.ToIntegerExpression())).
			ExecContext,
		r.tx)
}

func (r *UserRepository) SetPasswordResetToken(appUserId domain_model.AppUserID, token string, validUntil time.Time) error {
	return db.ExecSingleMutation(
		AppUser.UPDATE(AppUser.PasswordResetToken, AppUser.PasswordResetTokenValidUntil, AppUser.UpdatedAt).
			SET(token, validUntil, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserId.ToIntegerExpression())).
			ExecContext,
		r.tx)
}

func (r *UserRepository) SetPasswordHash(appUserId domain_model.AppUserID, hash string) error {
	return db.ExecSingleMutation(
		AppUser.UPDATE(AppUser.PasswordHash, AppUser.PasswordResetToken, AppUser.PasswordResetTokenValidUntil, AppUser.UpdatedAt).
			SET(hash, nil, nil, time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserId.ToIntegerExpression())).
			ExecContext,
		r.tx)
}

func (r *UserRepository) SetLanguage(appUserId domain_model.AppUserID, language domain_model.UserLanguage) error {
	return db.ExecSingleMutation(
		AppUser.UPDATE(AppUser.Language, AppUser.UpdatedAt).
			SET(model.UserLanguage(language), time.Now()).
			WHERE(AppUser.AppUserID.EQ(appUserId.ToIntegerExpression())).
			ExecContext,
		r.tx)
}

func (r *UserRepository) Insert(userRole domain_model.UserRole, userName string, email string, emailVerified bool, emailVerificationToken string, passwordHash string, language domain_model.UserLanguage) error {
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

	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()
	res := &model.AppUser{}
	err := stmt.QueryContext(ctx, r.tx, res)
	if err != nil {
		return errors.Wrap("cannot insert user", err)
	}

	return db.ExecSingleMutation(
		AppUserRole.INSERT(AppUserRole.Role, AppUserRole.AppUserID).
			VALUES(UserRole.User, res.AppUserID).
			ExecContext,
		r.tx)
}

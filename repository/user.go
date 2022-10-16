package repository

import (
	"context"
	"database/sql"
	"user-manager/db"
	"user-manager/db/generated/models"
	domain_model "user-manager/domain-model"
	"user-manager/util"
	"user-manager/util/nullable"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type UserRepository struct {
	tx *sql.Tx
}

func ProvideUserRepository(tx *sql.Tx) *UserRepository {
	return &UserRepository{tx}
}

func (r *UserRepository) GetUser(email string) (nullable.Nullable[*domain_model.AppUser], error) {
	user, err := db.Fetch(func(ctx context.Context) (*models.AppUser, error) {
		return models.AppUsers(
			models.AppUserWhere.Email.EQ(email),
			qm.Load(models.AppUserRels.AppUserRoles),
		).One(ctx, r.tx)
	})
	if err != nil {
		return nullable.Empty[*domain_model.AppUser](), util.Wrap("error fetching user", err)
	}
	if user.IsEmpty() {
		return nullable.Empty[*domain_model.AppUser](), nil
	}
	return nullable.NeverNil(domain_model.FromAppUserAndUserRolesModel(user.Val, user.Val.R.AppUserRoles)), nil
}

func (r *UserRepository) UpdateUserEmailVerificationToken(appUserId domain_model.AppUserID, token string) error {
	return db.ExecSingleMutation(func(ctx context.Context) (int64, error) {
		return (&models.AppUser{AppUserID: int64(appUserId), EmailVerificationToken: null.StringFrom(token)}).Update(ctx, r.tx, boil.Whitelist(models.AppUserColumns.EmailVerificationToken))
	})
}
func (r *UserRepository) UpdateUserEmailVerification(appUserId domain_model.AppUserID, token nullable.Nullable[string], verified bool) error {
	return db.ExecSingleMutation(func(ctx context.Context) (int64, error) {
		user := &models.AppUser{AppUserID: int64(appUserId), EmailVerified: verified}
		if token.IsPresent {
			user.EmailVerificationToken = null.StringFrom(token.Val)
		}
		return user.Update(ctx, r.tx, boil.Whitelist(models.AppUserColumns.EmailVerificationToken, models.AppUserColumns.EmailVerified))
	})
}

func (r *UserRepository) Insert(userRole domain_model.UserRole, userName string, email string, emailVerified bool, emailVerificationToken string, passwordHash string, language domain_model.UserLanguage) error {
	user := &models.AppUser{
		UserName:               userName,
		Email:                  email,
		EmailVerified:          emailVerified,
		EmailVerificationToken: null.StringFrom(emailVerificationToken),
		PasswordHash:           passwordHash,
		Language:               models.UserLanguage(language),
	}
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()
	if err := user.Insert(ctx, r.tx, boil.Infer()); err != nil {
		return util.Wrap("cannot insert user", err)
	}

	ctx, cancelTimeout = db.DefaultQueryContext()
	defer cancelTimeout()
	if err := user.AddAppUserRoles(ctx, r.tx, true, &models.AppUserRole{Role: models.UserRole(userRole)}); err != nil {
		return util.Wrap("cannot insert user role", err)
	}
	return nil
}

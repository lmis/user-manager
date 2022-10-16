package auth

// import (
// 	"context"
// 	ginext "user-manager/cmd/app/gin-extensions"
// 	auth_service "user-manager/cmd/app/services/auth"
// 	email_service "user-manager/cmd/app/services/email"
// 	"user-manager/db"
// 	"user-manager/db/generated/models"
// 	"user-manager/util"

// 	"github.com/gin-gonic/gin"
// 	"github.com/volatiletech/null/v8"
// 	"github.com/volatiletech/sqlboiler/v4/boil"
// )

// type SignUpTO struct {
// 	UserName string `json:"userName"`
// 	Language string `json:"language"`
// 	Email    string `json:"email"`
// 	Password []byte `json:"password"`
// }

// func PostSignUp(requestContext *ginext.RequestContext, requestTO *SignUpTO, _ *gin.Context) error {
// 	tx := requestContext.Tx
// 	securityLog := requestContext.SecurityLog

// 	maybeUser, err := db.Fetch(func(ctx context.Context) (*models.AppUser, error) {
// 		return models.AppUsers(models.AppUserWhere.Email.EQ(requestTO.Email)).One(ctx, tx)
// 	})
// 	if err != nil {
// 		return util.Wrap("error finding user", err)
// 	}
// 	if maybeUser.IsNotNil() {
// 		securityLog.Info("User already exists")
// 		if err = email_service.SendSignUpAttemptEmail(requestContext, maybeUser.Val); err != nil {
// 			return util.Wrap("error sending signup attempted email", err)
// 		}

// 		return nil
// 	}

// 	hash, err := auth_service.Hash(requestTO.Password)
// 	if err != nil {
// 		return util.Wrap("error hashing password", err)
// 	}

// 	language := models.UserLanguage(requestTO.Language)
// 	if language.IsValid() != nil {
// 		return util.Errorf("unsupported language \"%s\"", language.String())
// 	}
// 	user := &models.AppUser{
// 		UserName:               requestTO.UserName,
// 		Email:                  requestTO.Email,
// 		EmailVerified:          false,
// 		EmailVerificationToken: null.StringFrom(util.MakeRandomURLSafeB64(21)),
// 		PasswordHash:           hash,
// 		Language:               language,
// 	}

// 	ctx, cancelTimeout := db.DefaultQueryContext()
// 	defer cancelTimeout()
// 	if err = user.Insert(ctx, tx, boil.Infer()); err != nil {
// 		return util.Wrap("cannot insert user", err)
// 	}

// 	ctx, cancelTimeout = db.DefaultQueryContext()
// 	defer cancelTimeout()
// 	if err = user.AddAppUserRoles(ctx, tx, true, &models.AppUserRole{Role: "USER"}); err != nil {
// 		return util.Wrap("cannot insert user role", err)
// 	}

// 	if err = email_service.SendVerificationEmail(requestContext, user); err != nil {
// 		return util.Wrap("error sending verification email", err)
// 	}

// 	return nil
// }

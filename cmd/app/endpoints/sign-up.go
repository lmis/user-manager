package endpoints

import (
	"database/sql"
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	emailservice "user-manager/cmd/app/services/email"
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/crypto/bcrypt"
)

type SignUpTO struct {
	UserName string `json:"userName"`
	Language string `json:"language"`
	Email    string `json:"email"`
	Password []byte `json:"password"`
}

func PostSignUp(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	tx := requestContext.Tx
	securityLog := requestContext.SecurityLog
	var signUpTO SignUpTO
	if err := c.BindJSON(&signUpTO); err != nil {
		c.AbortWithError(http.StatusBadRequest, util.Wrap("cannot bind to signUpTO", err))
		return
	}

	var user *models.AppUser
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()

	user, err := models.AppUsers(models.AppUserWhere.Email.EQ(signUpTO.Email)).One(ctx, tx)
	if err != nil && err != sql.ErrNoRows {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("error finding user", err))
		return
	}
	if user != nil {
		securityLog.Info("User already exists")
		if err = emailservice.SendSignUpAttemptEmail(requestContext, user); err != nil {
			c.AbortWithError(http.StatusInternalServerError, util.Wrap("error sending signup attempted email", err))
			return
		}

		c.Status(http.StatusOK)
		return
	}

	hash, err := bcrypt.GenerateFromPassword(signUpTO.Password, bcrypt.DefaultCost)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("error hashing password", err))
		return
	}

	language := models.UserLanguage(signUpTO.Language)
	if language.IsValid() != nil {
		c.AbortWithError(http.StatusBadRequest, util.Errorf("unsupported language \"%s\"", language.String()))
		return
	}
	user = &models.AppUser{
		UserName:               signUpTO.UserName,
		Email:                  signUpTO.Email,
		Role:                   models.UserRoleUSER,
		EmailVerified:          false,
		EmailVerificationToken: null.StringFrom(util.MakeRandomURLSafeB64(21)),
		PasswordHash:           string(hash),
		Language:               language,
	}

	ctx, cancelTimeout = db.DefaultQueryContext()
	defer cancelTimeout()
	if err = user.Insert(ctx, tx, boil.Infer()); err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("cannot insert user", err))
		return
	}

	if err = emailservice.SendVerificationEmail(requestContext, user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("error sending verification email", err))
		return
	}

	c.Status(http.StatusOK)
}

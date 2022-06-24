package endpoints

import (
	"net/http"
	"user-manager/app/services"
	"user-manager/db"
	"user-manager/db/generated/models"
	ginext "user-manager/gin-extensions"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/crypto/bcrypt"
)

type SignUpTO struct {
	Email    string `json:"email"`
	Password []byte `json:"password"`
}

func PostSignUp(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	tx := requestContext.Tx
	securityLog := requestContext.SecurityLog
	var signUpTO SignUpTO
	err := c.BindJSON(&signUpTO)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, util.Wrap("PostSignUp", "cannot bind to signUpTO", err))
		return
	}

	var user *models.AppUser
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()

	user, err = models.AppUsers(models.AppUserWhere.Email.EQ(signUpTO.Email)).One(ctx, tx)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("PostSignup", "error finding user", err))
		return
	}
	if user != nil {
		securityLog.Info("User already exists")
		err = services.SendSignUpAttemptEmail(requestContext, user.Email)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, util.Wrap("PostSignup", "error sending signup attempted email", err))
			return
		}

		c.Status(http.StatusOK)
		return
	}

	hash, err := bcrypt.GenerateFromPassword(signUpTO.Password, bcrypt.DefaultCost)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("PostSignup", "error hashing password", err))
		return
	}
	user.PasswordHash = string(hash)

	user = &models.AppUser{
		PasswordHash:           string(hash),
		Email:                  signUpTO.Email,
		EmailVerificationToken: null.StringFrom(util.MakeRandomURLSafeB64(21)),
		Role:                   models.UserRoleUSER,
		Verified:               false,
	}

	ctx, cancelTimeout = db.DefaultQueryContext()
	defer cancelTimeout()
	err = user.Insert(ctx, tx, boil.Infer())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("PostSignup", "cannot insert user", err))
		return
	}

	err = services.SendVerificationEmail(requestContext, user.Email)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("PostSignup", "error sending verification email", err))
		return
	}

	c.Status(http.StatusOK)
}

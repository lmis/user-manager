package resource

import (
	ginext "user-manager/cmd/app/gin-extensions"
	domain_model "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/util"
	"user-manager/util/nullable"

	"github.com/gin-gonic/gin"
)

type EmailConfirmationResource struct {
	securityLog    domain_model.SecurityLog
	userSession    nullable.Nullable[*domain_model.UserSession]
	userRepository *repository.UserRepository
}

func ProvideEmailConfirmationResource(securityLog domain_model.SecurityLog, userSession nullable.Nullable[*domain_model.UserSession], userRepository *repository.UserRepository) *EmailConfirmationResource {
	return &EmailConfirmationResource{securityLog, userSession, userRepository}
}

func RegisterEmailConfirmationResource(group *gin.RouterGroup) {
	group.POST("confirm-email", ginext.WrapEndpoint(InitializeEmailConfirmationResource, (*EmailConfirmationResource).Post))
}

type EmailConfirmationTO struct {
	Token string `json:"token"`
}

type EmailConfirmationStatus string

const (
	AlreadyConfirmed EmailConfirmationStatus = "already-confirmed"
	NewlyConfirmed   EmailConfirmationStatus = "newly-confirmed"
	InvalidToken     EmailConfirmationStatus = "invalid-token"
)

type EmailConfirmationResponseTO struct {
	Status EmailConfirmationStatus `json:"status"`
}

func (r *EmailConfirmationResource) Post(request *EmailConfirmationTO) (*EmailConfirmationResponseTO, error) {
	securityLog := r.securityLog
	user := r.userSession.OrPanic().User

	if user.EmailVerified {
		securityLog.Info("Email already verified")
		return &EmailConfirmationResponseTO{
			Status: AlreadyConfirmed,
		}, nil
	}

	if user.EmailVerificationToken.IsEmpty() {
		return nil, util.Error("no verification token present on database")
	}

	if request.Token != user.EmailVerificationToken.Val {
		securityLog.Info("Invalid email verification token")
		return &EmailConfirmationResponseTO{
			Status: InvalidToken,
		}, nil
	}

	if err := r.userRepository.UpdateUserEmailVerification(user.AppUserID, nullable.Empty[string](), true); err != nil {
		return nil, util.Wrap("issue updating user email verficiation", err)
	}

	return &EmailConfirmationResponseTO{
		Status: NewlyConfirmed,
	}, nil
}

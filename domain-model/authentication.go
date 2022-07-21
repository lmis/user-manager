package domain_model

import (
	"user-manager/db/generated/models"
)

type Authentication struct {
	UserSession *models.UserSession
	AppUser     *models.AppUser
}

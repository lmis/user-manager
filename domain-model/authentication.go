package domain_model

import (
	"user-manager/db/generated/models"
)

type Authentication struct {
	UserRoles   []models.UserRole
	UserSession *models.UserSession
	AppUser     *models.AppUser
}

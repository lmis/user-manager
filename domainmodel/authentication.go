package domainmodel

import (
	"user-manager/db/generated/models"
)

type Authentication struct {
	UserID      int
	Role        models.UserRole
	UserSession *models.UserSession
}

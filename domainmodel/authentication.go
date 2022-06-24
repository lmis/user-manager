package domainmodel

import (
	"user-manager/db/generated/models"
	appuser "user-manager/domainmodel/id/appUser"
)

type Authentication struct {
	UserID        appuser.ID
	Role          models.UserRole
	EmailVerified bool
	UserSession   *models.UserSession
}

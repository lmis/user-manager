package domain_model

import (
	"user-manager/db/generated/models"
)

type UserRole models.UserRole

const (
	USER_ROLE_USER        UserRole = UserRole(models.UserRoleUSER)
	USER_ROLE_ADMIN       UserRole = UserRole(models.UserRoleADMIN)
	USER_ROLE_SUPER_ADMIN UserRole = UserRole(models.UserRoleSUPER_ADMIN)
)

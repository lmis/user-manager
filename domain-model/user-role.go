package domain_model

import (
	"user-manager/db/generated/models/postgres/public/model"
)

type UserRole model.UserRole

const (
	UserRoleUser       = UserRole(model.UserRole_User)
	UserRoleAdmin      = UserRole(model.UserRole_Admin)
	UserRoleSuperAdmin = UserRole(model.UserRole_SuperAdmin)
)

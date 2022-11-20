package domain_model

import (
	"user-manager/db/generated/models/postgres/public/model"
)

type UserRole model.UserRole

const (
	USER_ROLE_USER        UserRole = UserRole(model.UserRole_User)
	USER_ROLE_ADMIN       UserRole = UserRole(model.UserRole_Admin)
	USER_ROLE_SUPER_ADMIN UserRole = UserRole(model.UserRole_SuperAdmin)
)

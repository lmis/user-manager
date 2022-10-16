package domain_model

import (
	"user-manager/db/generated/models"
	"user-manager/util/slices"
)

type UserLanguage models.UserLanguage

func (l UserLanguage) IsValid() bool {
	return models.UserLanguage(l).IsValid() == nil
}

func AllUserLanguage() []UserLanguage {
	return slices.Map(models.AllUserLanguage(), func(m models.UserLanguage) UserLanguage { return UserLanguage(m) })
}

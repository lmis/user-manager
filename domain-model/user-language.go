package domain_model

import (
	"user-manager/db/generated/models/postgres/public/model"
	"user-manager/util/slices"
)

type UserLanguage model.UserLanguage

func (l UserLanguage) IsValid() bool {
	return slices.Contains(AllUserLanguages(), l)
}

func AllUserLanguages() []UserLanguage {
	return []UserLanguage{UserLanguage(model.UserLanguage_En), UserLanguage(model.UserLanguage_De)}
}

package injector

import (
	"embed"
	"fmt"
	"text/template"
	domain_model "user-manager/domain-model"
	"user-manager/util/errors"

	"gopkg.in/yaml.v2"
)

var baseTemplate *template.Template
var translations map[domain_model.UserLanguage]domain_model.Translations = make(map[domain_model.UserLanguage]domain_model.Translations)

func SetupEmailTemplatesProviders(translationsFS embed.FS) error {
	t, err := template.New("base").Parse(`
	{{.Salutation}}

	{{range .Paragraphs}}
		{{- .}}
	{{end}}

	{{.Footer}}
	`)
	if err != nil {
		return errors.Wrap("base template cannot be parsed", err)
	}
	baseTemplate = t

	for _, lang := range domain_model.AllUserLanguages() {
		file, err := translationsFS.ReadFile(fmt.Sprintf("translations/%s.yaml", lang))
		if err != nil {
			return errors.Wrap(fmt.Sprintf("%s translations cannot be read", lang), err)
		}
		translation := domain_model.Translations{}
		if err = yaml.Unmarshal(file, &translation); err != nil {
			return errors.Wrap(fmt.Sprintf("%s translations cannot be parsed", lang), err)
		}
		translations[lang] = translation
	}
	return nil
}

func ProvideBaseTemplate() *template.Template {
	return baseTemplate
}

func ProvideTranslations() map[domain_model.UserLanguage]domain_model.Translations {
	return translations
}

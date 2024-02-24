package middleware

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"text/template"
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
	"user-manager/util/errors"
)

func RegisterRequestContextMiddleware(app *gin.Engine, translationsFS embed.FS, config *dm.Config) error {
	if config == nil {
		return errors.Error("Invalid config: nil")
	}
	emailing := dm.Emailing{Translations: make(map[dm.UserLanguage]dm.Translations)}
	// TODO: Embed as template
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
	emailing.BaseTemplate = t

	for _, lang := range dm.AllUserLanguages() {
		file, err := translationsFS.ReadFile(fmt.Sprintf("translations/%s.yaml", lang))
		if err != nil {
			return errors.Wrap(fmt.Sprintf("%s translations cannot be read", lang), err)
		}
		translation := dm.Translations{}
		if err = yaml.Unmarshal(file, &translation); err != nil {
			return errors.Wrap(fmt.Sprintf("%s translations cannot be parsed", lang), err)
		}
		emailing.Translations[lang] = translation
	}
	app.Use(func(ctx *gin.Context) {
		ctx.Set(ginext.RequestContextKey, &ginext.RequestContext{Emailing: emailing, Config: config})
	})
	return nil
}

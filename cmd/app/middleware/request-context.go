package middleware

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/yaml.v3"
	"text/template"
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
)

func RegisterRequestContextMiddleware(app *gin.Engine, translationsFS embed.FS, database *mongo.Database, config *dm.Config) error {
	if config == nil {
		return errs.Error("Invalid config: nil")
	}
	if database == nil {
		return errs.Error("Invalid database: nil")
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
		return errs.Wrap("base template cannot be parsed", err)
	}
	emailing.BaseTemplate = t

	for _, lang := range dm.AllUserLanguages() {
		file, err := translationsFS.ReadFile(fmt.Sprintf("translations/%s.yaml", lang))
		if err != nil {
			return errs.Wrap(fmt.Sprintf("%s translations cannot be read", lang), err)
		}
		translation := dm.Translations{}
		if err = yaml.Unmarshal(file, &translation); err != nil {
			return errs.Wrap(fmt.Sprintf("%s translations cannot be parsed", lang), err)
		}
		emailing.Translations[lang] = translation
	}
	app.Use(func(ctx *gin.Context) {
		ctx.Set(ginext.RequestContextKey, &ginext.RequestContext{Emailing: emailing, Config: config, Database: database})
	})
	return nil
}

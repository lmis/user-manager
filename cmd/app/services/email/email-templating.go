package emailservice

import (
	"embed"
	"fmt"
	"strings"
	"text/template"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/util"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"gopkg.in/yaml.v2"
)

var baseTemplate *template.Template
var translations map[models.UserLanguage]translation = make(map[models.UserLanguage]translation)

type Priority int

const (
	PriorityLow  Priority = 10
	PriorityMid  Priority = 100
	PriorityHigh Priority = 1000
)

type baseTemplateData struct {
	Salutation string
	Paragraphs []string
	Footer     string
}

type translation struct {
	Salutation          string   `yaml:"salutation"`
	SalutationAnonymous string   `yaml:"salutationAnonymous"`
	VerificationEmail   []string `yaml:"verificationEmail"`
	SignUpAttemptEmail  []string `yaml:"signUpAttemptEmail"`
	Footer              string   `yaml:"footer"`
}

func Initialize(log util.Logger, translationsFS embed.FS) error {
	t, err := template.New("base").Parse(`
	{{.Salutation}}

	{{range .Paragraphs}}
		{{- .}}
	{{end}}

	{{.Footer}}
	`)
	if err != nil {
		return util.Wrap("base template cannot be parsed", err)
	}
	baseTemplate = t

	for _, lang := range models.AllUserLanguage() {
		file, err := translationsFS.ReadFile(fmt.Sprintf("translations/%s.yaml", lang))
		if err != nil {
			return util.Wrap(fmt.Sprintf("%s translations cannot be read", lang), err)
		}
		translation := translation{}
		if err = yaml.Unmarshal(file, &translation); err != nil {
			return util.Wrap(fmt.Sprintf("%s translations cannot be parsed", lang), err)
		}
		translations[lang] = translation
	}
	return nil
}

func executeTemplate(templateText string, data interface{}) (string, error) {
	t, err := template.New("base").Parse(templateText)
	if err != nil {
		return "", util.Wrap("templateText canot be parsed", err)
	}
	writer := &strings.Builder{}
	if err = t.Execute(writer, data); err != nil {
		return "", util.Wrap("template execution error", err)
	}
	return writer.String(), nil
}

func enqueueBasicEmail(r *ginext.RequestContext,
	language models.UserLanguage,
	templates []string,
	data map[string]string,
	from string,
	to string,
	priority Priority,
) error {
	var err error
	translation := translations[language]
	salutation := ""
	_, ok := data["UserName"]
	if ok {
		salutation, err = executeTemplate(translation.Salutation, data)
		if err != nil {
			return util.Wrap("issue translating salutation", err)
		}
	} else {
		salutation = translation.SalutationAnonymous
	}
	subject := ""
	paragraphs := []string{}
	for i, t := range templates {
		p, err := executeTemplate(t, data)
		if err != nil {
			return util.Wrap(fmt.Sprintf("issue translating template %d", i), err)
		}
		if i == 0 {
			subject = p
		} else {
			paragraphs = append(paragraphs, p)
		}
	}
	writer := &strings.Builder{}
	if err = baseTemplate.Execute(writer, baseTemplateData{
		Salutation: salutation,
		Paragraphs: paragraphs,
		Footer:     translation.Footer,
	}); err != nil {
		return util.Wrap("issue executing base template", err)
	}

	tx := r.Tx
	mail := models.MailQueue{
		FromAddress: from,
		ToAddress:   to,
		Content:     writer.String(),
		Subject:     subject,
		Status:      models.EmailStatusPENDING,
		Priority:    int16(priority),
	}
	ctx, cancel := db.DefaultQueryContext()
	defer cancel()
	if err = mail.Insert(ctx, tx, boil.Infer()); err != nil {
		return util.Wrap("issue inserting email in db", err)
	}

	return nil
}

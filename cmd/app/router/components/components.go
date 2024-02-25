package components

import (
	"embed"
	"html/template"
	"io"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
)

const templatesPattern = "templates/*"

type TemplateData struct {
	ServiceName string
}

//go:embed templates/layout.tmpl
var layoutFS embed.FS
var layoutTemplate *template.Template

func init() {
	layoutTemplate = template.Must(template.New("layout.tmpl").ParseFS(layoutFS, templatesPattern))

	indexTemplate = template.Must(template.Must(layoutTemplate.Clone()).ParseFS(indexFS, templatesPattern))
}

//go:embed templates/index.tmpl
var indexFS embed.FS
var indexTemplate *template.Template

func Index(r *dm.RequestContext, writer io.Writer) error {
	config := r.Config

	err := indexTemplate.ExecuteTemplate(writer, "layout.tmpl", TemplateData{
		ServiceName: config.ServiceName,
	})
	if err != nil {
		return errs.Wrap("error executing subject template", err)
	}

	return nil
}

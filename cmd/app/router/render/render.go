package render

import (
	"github.com/a-h/templ"
	dm "user-manager/domain-model"
	"user-manager/util/random"
)

func Index(r *dm.RequestContext) (templ.Component, error) {
	config := r.Config

	return indexPage(config.ServiceName, "Hello", random.MakeRandomURLSafeB64(21), random.MakeRandomURLSafeB64(21), config.Environment), nil
}

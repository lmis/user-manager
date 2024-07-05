package render

import (
	"fmt"
	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"user-manager/cmd/app/router/render/layout"
	"user-manager/cmd/app/router/render/user"
	dm "user-manager/domain-model"
	"user-manager/util/random"
)

// TODO: Move to 'resource' package (also rename that package)
func UserHome(c *gin.Context, r *dm.RequestContext) (templ.Component, error) {
	config := r.Config

	return FullPage(c, r, "Home", user.Home(config.ServiceName)), nil
}

func FullPage(ctx *gin.Context, r *dm.RequestContext, title string, component templ.Component) templ.Component {
	config := r.Config

	nonce := random.MakeRandomURLSafeB64(21)
	// TODO: This severly slows down the pageload.
	ctx.Header("Content-Security-Policy", fmt.Sprintf(`
           default-src 'self';
           script-src 'strict-dynamic' 'nonce-%s' 'unsafe-inline';
           object-src 'none';
           base-uri 'none';
           connect-src 'self';
           style-src 'self' 'sha256-d7rFBVhb3n/Drrf+EpNWYdITkos3kQRFpB0oSOycXg4=' 'sha256-bsV5JivYxvGywDAZ22EZJKBFip65Ng9xoJVLbBg7bdo=';
       `, nonce))
	return layout.Page(config.ServiceName, title, nonce, random.MakeRandomURLSafeB64(21), config.Environment, func() templ.Component { return component })
}

package ginext

import "github.com/gin-gonic/gin"

func HXLocationOrRedirect(c *gin.Context, location string) {
	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Location", location)
		c.Status(204)
	} else {
		c.Redirect(302, location)
	}
}

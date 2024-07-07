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

func HXIsFullPageLoad(c *gin.Context) bool {
	return c.GetHeader("HX-Request") != "true" || c.GetHeader("HX-History-Restore-Request") == "true"
}
func HXReswap(c *gin.Context, value string) {
	c.Header("HX-Reswap", value)
}

func HXRetarget(c *gin.Context, value string) {
	c.Header("HX-Retarget", value)
}
func HXReload(c *gin.Context) {
	c.Header("HX-Location", c.GetHeader("HX-Current-URL"))
}

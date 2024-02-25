package ginext

import (
	"net/http"
	"user-manager/cmd/app/router/middleware"
	dm "user-manager/domain-model"
	"user-manager/util/errs"

	"github.com/gin-gonic/gin"
)

type Endpoint[requestTO interface{}, responseTO interface{}] func(*gin.Context, *dm.RequestContext, requestTO) (responseTO, error)
type EndpointWithoutResponseBody[requestTO interface{}] func(*gin.Context, *dm.RequestContext, requestTO) error
type EndpointWithoutRequestBody[responseTO interface{}] func(*gin.Context, *dm.RequestContext) (responseTO, error)
type EndpointWithoutRequestOrResponseBody func(*gin.Context, *dm.RequestContext) error

func WrapEndpoint[requestTO interface{}, responseTO interface{}](endpoint Endpoint[requestTO, responseTO]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request requestTO
		if err := c.BindJSON(&request); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, errs.Wrap("cannot bind to request TO", err))
			return
		}
		response, err := endpoint(c, middleware.GetRequestContext(c), request)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if !c.IsAborted() {
			c.JSON(http.StatusOK, response)
		}
	}
}

func WrapEndpointWithoutResponseBody[requestTO interface{}](endpoint EndpointWithoutResponseBody[requestTO]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request requestTO
		if err := c.BindJSON(&request); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, errs.Wrap("cannot bind to request TO", err))
			return
		}
		if err := endpoint(c, middleware.GetRequestContext(c), request); err != nil {
			status := http.StatusInternalServerError
			_ = c.AbortWithError(status, err)
			return
		}

		if !c.IsAborted() {
			c.Status(http.StatusNoContent)
		}
	}
}

func WrapEndpointWithoutRequestBody[responseTO interface{}](endpoint EndpointWithoutRequestBody[responseTO]) gin.HandlerFunc {
	return func(c *gin.Context) {
		response, err := endpoint(c, middleware.GetRequestContext(c))
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if !c.IsAborted() {
			c.JSON(http.StatusOK, response)
		}
	}
}

func WrapEndpointWithoutRequestOrResponseBody(endpoint EndpointWithoutRequestOrResponseBody) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := endpoint(c, middleware.GetRequestContext(c)); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if !c.IsAborted() {
			c.Status(http.StatusNoContent)
		}
	}
}

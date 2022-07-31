package ginext

import (
	"net/http"
	"user-manager/util"

	"github.com/gin-gonic/gin"
)

type Endpoint[requestTO interface{}, responseTO interface{}] func(requestContext *RequestContext, request requestTO, c *gin.Context) (responseTO, error)
type EndpointWithoutResponseBody[requestTO interface{}] func(requestContext *RequestContext, request requestTO, c *gin.Context) error
type EndpointWithoutRequestBody[responseTO interface{}] func(requestContext *RequestContext, c *gin.Context) (responseTO, error)
type EndpointWithoutRequestOrResponseBody func(requestContext *RequestContext, c *gin.Context) error

func WrapEndpoint[requestTO interface{}, responseTO interface{}](endpoint Endpoint[requestTO, responseTO]) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestContext := GetRequestContext(c)
		var request requestTO
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithError(http.StatusBadRequest, util.Wrap("cannot bind to request TO", err))
			return
		}
		response, err := endpoint(requestContext, request, c)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

func WrapEndpointWithoutResponseBody[requestTO interface{}](endpoint EndpointWithoutResponseBody[requestTO]) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestContext := GetRequestContext(c)
		var request requestTO
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithError(http.StatusBadRequest, util.Wrap("cannot bind to request TO", err))
			return
		}
		if err := endpoint(requestContext, request, c); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func WrapEndpointWithoutRequestBody[responseTO interface{}](endpoint EndpointWithoutRequestBody[responseTO]) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestContext := GetRequestContext(c)
		response, err := endpoint(requestContext, c)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

func WrapEndpointWithoutRequestOrResponseBody(endpoint EndpointWithoutRequestOrResponseBody) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestContext := GetRequestContext(c)
		if err := endpoint(requestContext, c); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusNoContent)
	}
}

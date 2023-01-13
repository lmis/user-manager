package ginext

import (
	"net/http"
	"user-manager/util/errors"

	"github.com/gin-gonic/gin"
)

type ResourceInitializer[T interface{}] func(*gin.Context) T

type Endpoint[T interface{}, requestTO interface{}, responseTO interface{}] func(T, requestTO) (responseTO, error)
type EndpointWithoutResponseBody[T interface{}, requestTO interface{}] func(T, requestTO) error
type EndpointWithoutRequestBody[T interface{}, responseTO interface{}] func(T) (responseTO, error)
type EndpointWithoutRequestOrResponseBody[T interface{}] func(T) error

func WrapEndpoint[T interface{}, requestTO interface{}, responseTO interface{}](initialzeResource ResourceInitializer[T], endpoint Endpoint[T, requestTO, responseTO]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request requestTO
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithError(http.StatusBadRequest, errors.Wrap("cannot bind to request TO", err))
			return
		}
		response, err := endpoint(initialzeResource(c), request)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

func WrapEndpointWithoutResponseBody[T interface{}, requestTO interface{}](initializeResource ResourceInitializer[T], endpoint EndpointWithoutResponseBody[T, requestTO]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request requestTO
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithError(http.StatusBadRequest, errors.Wrap("cannot bind to request TO", err))
			return
		}
		if err := endpoint(initializeResource(c), request); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func WrapEndpointWithoutRequestBody[T interface{}, responseTO interface{}](initializeResource ResourceInitializer[T], endpoint EndpointWithoutRequestBody[T, responseTO]) gin.HandlerFunc {
	return func(c *gin.Context) {
		response, err := endpoint(initializeResource(c))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

func WrapEndpointWithoutRequestOrResponseBody[T interface{}](initializeResource ResourceInitializer[T], endpoint EndpointWithoutRequestOrResponseBody[T]) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := endpoint(initializeResource(c)); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusNoContent)
	}
}

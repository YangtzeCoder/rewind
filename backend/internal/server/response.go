package server

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

const (
	errInvalidArgument    = "INVALID_ARGUMENT"
	errInvalidCredentials = "INVALID_CREDENTIALS"
	errUnauthorized       = "UNAUTHORIZED"
	errNotFound           = "NOT_FOUND"
	errMethodNotAllowed   = "METHOD_NOT_ALLOWED"
	errInternal           = "INTERNAL_ERROR"
)

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiResponse struct {
	Success bool      `json:"success"`
	Data    any       `json:"data,omitempty"`
	Error   *apiError `json:"error,omitempty"`
}

// success omits data when nil; typed nil slices are returned as empty arrays.
func success(c *gin.Context, data any) {
	if data != nil {
		v := reflect.ValueOf(data)
		if v.Kind() == reflect.Slice && v.IsNil() {
			data = reflect.MakeSlice(v.Type(), 0, 0).Interface()
		}
	}
	c.JSON(http.StatusOK, apiResponse{Success: true, Data: data})
}

func failure(c *gin.Context, code, message string) {
	c.JSON(http.StatusOK, apiResponse{Success: false, Error: &apiError{Code: code, Message: message}})
}

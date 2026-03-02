package openapi

import (
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

var DefaultErrorCodes = []int{
	http.StatusBadRequest,
	http.StatusUnauthorized,
	http.StatusForbidden,
	http.StatusNotFound,
	http.StatusUnprocessableEntity,
	http.StatusInternalServerError,
}

func ErrorCodes(codes ...int) []int {
	return append(DefaultErrorCodes, codes...)
}

func ErrorBadRequest(msg string, errs ...error) StatusError {
	return huma.Error400BadRequest(msg, errs...)
}

func ErrorUnauthorized(msg string, errs ...error) StatusError {
	return huma.Error401Unauthorized(msg, errs...)
}

func ErrorForbidden(msg string, errs ...error) StatusError {
	return huma.Error403Forbidden(msg, errs...)
}

func ErrorNotFound(msg string, errs ...error) StatusError {
	return huma.Error404NotFound(msg, errs...)
}

func ErrorInternal(msg string, errs ...error) StatusError {
	return huma.Error500InternalServerError(msg, errs...)
}

func IsClientError(err error) bool {
	var statusError StatusError
	if errors.As(err, &statusError) {
		if statusError.GetStatus() < 500 {
			return true //, statusError.GetStatus()
		}
	}
	return false
}

func NewErrorDetail(msg string, location string, val any) *ErrorDetail {
	return &ErrorDetail{
		Message:  msg,
		Location: location,
		Value:    val,
	}
}

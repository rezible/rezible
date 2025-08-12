package openapi

import (
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

var defaultErrorCodes = []int{
	http.StatusBadRequest,
	http.StatusUnauthorized,
	http.StatusForbidden,
	http.StatusNotFound,
	http.StatusInternalServerError,
}

func ErrorBadRequest(msg string, errs ...error) huma.StatusError {
	return huma.Error400BadRequest(msg, errs...)
}

func ErrorUnauthorized(msg string, errs ...error) huma.StatusError {
	return huma.Error401Unauthorized(msg, errs...)
}

func ErrorNotFound(msg string, errs ...error) huma.StatusError {
	return huma.Error404NotFound(msg, errs...)
}

func ErrorForbidden(msg string, errs ...error) huma.StatusError {
	return huma.Error403Forbidden(msg, errs...)
}

func ErrorInternal(msg string, errs ...error) huma.StatusError {
	return huma.Error500InternalServerError(msg, errs...)
}

func IsClientError(err error) bool {
	var statusError huma.StatusError
	if errors.As(err, &statusError) {
		if statusError.GetStatus() < 500 {
			return true //, statusError.GetStatus()
		}
	}
	return false
}

func errorCodes(codes ...int) []int {
	return append(defaultErrorCodes, codes...)
}

func NewErrorDetail(msg string, location string, val any) *huma.ErrorDetail {
	return &huma.ErrorDetail{
		Message:  msg,
		Location: location,
		Value:    val,
	}
}

func interceptErrors(s Handler) func(huma.Context, string, any) (any, error) {
	return func(ctx huma.Context, status string, v any) (any, error) {
		//if err, ok := v.(error); ok {
		//	log.Debug().Str("path", ctx.URL().Path).Err(err).Msg("error!")
		//}
		//if err, ok := v.(error); ok {
		//	if newErr := s.InterceptError(ctx, 0, err); newErr != nil {
		//		v = newErr
		//	}
		//}
		/*
			if err, ok := v.(error); ok {
				var mdl *huma.ErrorModel
				if errors.As(err, &mdl) {
					//for _, e := range mdl.Errors {
					//	err = errors.Join(err, e)
					//}
					statusCode = mdl.GetStatus()
					// mdl.Errors = nil
				} else {
					code, statusErr := strconv.Atoi(status)
					if statusErr != nil {
						log.Error().Err(statusErr).Str("status", status).Msg("failed to convert status code")
						return v, nil
					}
					statusCode = code
				}
				var statusCode int
				if newErr := s.InterceptError(ctx, statusCode, err); newErr != nil {
					v = newErr
				}
			}
		*/
		return v, nil
	}
}

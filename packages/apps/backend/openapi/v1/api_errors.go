package v1

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
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

func NewErrorDetail(msg string, location string, val any) *huma.ErrorDetail {
	return &huma.ErrorDetail{
		Message:  msg,
		Location: location,
		Value:    val,
	}
}

var (
	uniqueErrFieldRe         = regexp.MustCompile("unique constraint \".*_(.*)_key\"")
	enumValidationErrFieldRe = regexp.MustCompile("invalid enum value for")

	commonConstraints = map[string]string{
		"name":  "Name already exists",
		"value": "Value already exists",
	}
)

func isClientError(err error) bool {

	return false
}

func asStatusError(msg string, err error) huma.StatusError {
	var statusError huma.StatusError
	if errors.As(err, &statusError) {
		return statusError
	}

	if ent.IsNotFound(err) {
		return huma.Error404NotFound("not found", err)
	}

	if enumValidationErrFieldRe.MatchString(err.Error()) {
		return huma.Error400BadRequest("validation error", err)
	}

	if ent.IsConstraintError(err) {
		match := uniqueErrFieldRe.FindStringSubmatch(err.Error())
		if match == nil || len(match) < 2 {
			return huma.Error400BadRequest("Constraint failed")
		}

		field := match[1]
		cstrMsg, found := commonConstraints[field]
		if found {
			return huma.Error400BadRequest("Constraint Error", NewErrorDetail(cstrMsg, field, nil))
		}
		return huma.Error400BadRequest("Value is not unique")
	}

	return huma.Error500InternalServerError(msg, err)
}

func Error(msg string, err error) error {
	statusErr := asStatusError(msg, err)

	logEvent := log.Warn()
	if statusErr.GetStatus() >= 500 {
		logEvent = log.Error()
	}
	logEvent.
		Str("message", msg).
		Int("status", statusErr.GetStatus()).
		AnErr("error", err).
		Msg("API Error")

	return statusErr
}

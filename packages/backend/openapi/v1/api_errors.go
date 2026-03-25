package v1

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rezible/rezible/ent"
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

func isClientError(err error) bool {
	var statusError huma.StatusError
	if errors.As(err, &statusError) {
		if statusError.GetStatus() < 500 {
			return true
		}
	}
	return false
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

func Error(msg string, err error) error {
	if isClientError(err) {
		return err
	}

	if ent.IsNotFound(err) {
		return huma.Error404NotFound("Not found")
	}

	if enumValidationErrFieldRe.MatchString(err.Error()) {
		return err
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

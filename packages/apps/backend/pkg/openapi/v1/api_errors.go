package v1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/danielgtaylor/huma/v2"
	rez "github.com/rezible/rezible"
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

type statusErrorFunc = func(...error) huma.StatusError

func statusErrorWithCode(code string, fn func(string, ...error) huma.StatusError) statusErrorFunc {
	return func(errs ...error) huma.StatusError {
		return fn(code, errs...)
	}
}

var (
	Err400InvalidInput = statusErrorWithCode("invalid_input", huma.Error400BadRequest)

	Err401AuthSessionMissing = statusErrorWithCode("auth_session_missing", huma.Error401Unauthorized)
	Err401AuthSessionExpired = statusErrorWithCode("auth_session_expired", huma.Error401Unauthorized)
	Err401AuthSessionInvalid = statusErrorWithCode("auth_session_invalid", huma.Error401Unauthorized)

	Err403DomainNotAllowed = statusErrorWithCode("domain_not_allowed", huma.Error403Forbidden)
	Err403Forbidden        = statusErrorWithCode("forbidden", huma.Error403Forbidden)

	Err404NotFound = statusErrorWithCode("not_found", huma.Error404NotFound)

	Err409Conflict = statusErrorWithCode("conflict", huma.Error409Conflict)

	Err500Internal = statusErrorWithCode("internal_server_error", huma.Error500InternalServerError)

	Err501NotImplemented = statusErrorWithCode("not_implemented", huma.Error501NotImplemented)

	// TODO: just have one map?

	requestErrorMap = map[error]statusErrorFunc{
		rez.ErrInvalidInput:       Err400InvalidInput,
		rez.ErrAuthSessionMissing: Err401AuthSessionMissing,
		rez.ErrAuthSessionExpired: Err401AuthSessionExpired,
		rez.ErrAuthSessionInvalid: Err401AuthSessionInvalid,
		rez.ErrInvalidUser:        Err401AuthSessionInvalid,
		rez.ErrInvalidTenant:      Err401AuthSessionInvalid,
		rez.ErrDomainNotAllowed:   Err403DomainNotAllowed,
		rez.ErrForbidden:          Err403Forbidden,
		rez.ErrNotFound:           Err404NotFound,
		rez.ErrConflict:           Err409Conflict,
		rez.ErrNotImplemented:     Err501NotImplemented,
	}
)

func Error(ctx context.Context, msg string, err error) error {
	statusErr := ConvertStatusError(msg, err)

	logLevel := slog.LevelWarn
	if statusErr.GetStatus() >= 500 {
		logLevel = slog.LevelError
	}
	slog.Log(ctx, logLevel, "API Error",
		"message", msg,
		"status", statusErr.GetStatus(),
		"error", err,
	)

	return statusErr
}

func ConvertStatusError(msg string, err error) huma.StatusError {
	if statusError, isStatusErr := errors.AsType[huma.StatusError](err); isStatusErr {
		return statusError
	}

	if ent.IsNotFound(err) {
		return Err404NotFound()
	}

	if enumValidationErrFieldRe.MatchString(err.Error()) {
		return Err400InvalidInput()
	}

	if ent.IsConstraintError(err) {
		match := uniqueErrFieldRe.FindStringSubmatch(err.Error())
		if match != nil && len(match) > 2 {
			field := match[1]
			if constraintMsg, found := commonConstraints[field]; found {
				return Err400InvalidInput(NewErrorDetail(constraintMsg, field, nil))
			}
			return Err409Conflict()
		}
		return Err400InvalidInput()
	}

	for domainErr, statusErrFn := range requestErrorMap {
		if errors.Is(err, domainErr) {
			return statusErrFn(err)
		}
	}

	slog.Error("unexpected auth status error",
		"msg", msg,
		"error", err)
	return Err500Internal(err)
}

package v1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

// errorModel is the body of every error response. Code lets callers branch without parsing messages.
type errorModel struct {
	huma.ErrorModel
	Code errs.Code `json:"code" doc:"A stable code for the kind of error."`

	// logged is set once the error is logged and recorded on the span. Errors reach a response by two
	// paths: Error, which knows the handler's message, and Huma's own write path. Huma has no single hook
	// that sees every error, so this keeps an error that went through both from being logged twice.
	logged bool
}

func (o operations) RegisterErrorEnums(api huma.API) {
	registerEnumAlias[errs.Code, errorCodeSchema](api)
}

type errorCodeSchema errs.Code

func (errorCodeSchema) Schema(huma.Registry) *huma.Schema {
	codes := errs.Codes()
	values := make([]string, len(codes))
	for i, code := range codes {
		values[i] = string(code)
	}
	return makeEnumStringSchema(values)
}

var codeStatuses = map[errs.Code]int{
	errs.CodeInvalidInput:    http.StatusBadRequest,
	errs.CodeUnprocessable:   http.StatusUnprocessableEntity,
	errs.CodeUnauthenticated: http.StatusUnauthorized,
	errs.CodeForbidden:       http.StatusForbidden,
	errs.CodeNotFound:        http.StatusNotFound,
	errs.CodeConflict:        http.StatusConflict,
	errs.CodeRateLimited:     http.StatusTooManyRequests,
	errs.CodeUnavailable:     http.StatusServiceUnavailable,
	errs.CodeNotImplemented:  http.StatusNotImplemented,
	errs.CodeInternal:        http.StatusInternalServerError,
}

// init replaces Huma's error constructors for every importer of this package, including the spec
// generator, as Huma intends, so every error response is an errorModel.
func init() {
	huma.NewError = func(status int, msg string, errList ...error) huma.StatusError {
		return newError(status, msg, errList...)
	}
	huma.NewErrorWithContext = newErrorWithContext
}

// newErrorWithContext builds and logs the errors Huma writes, such as request validation failures. An
// errorModel passed to huma.WriteErr is written as it is, and logged unless it already was.
func newErrorWithContext(hctx huma.Context, status int, msg string, errList ...error) huma.StatusError {
	var causes []error
	for _, e := range errList {
		if model, isModel := e.(*errorModel); isModel {
			if !model.logged {
				recordError(hctx.Context(), msg, model, nil)
			}
			return model
		}
		if _, isDetailer := e.(huma.ErrorDetailer); !isDetailer && e != nil {
			causes = append(causes, e)
		}
	}
	model := newError(status, msg, errList...)
	recordError(hctx.Context(), msg, model, errors.Join(causes...))
	return model
}

// newError builds every error response, including Huma's own. Only Huma error details reach the
// response; any other error passed is read for its code and never written. The code is that of the
// first *errs.Error passed, or else the status's. A 5xx never shows msg, only its code's default text.
func newError(status int, msg string, errList ...error) *errorModel {
	var code errs.Code
	var details []*huma.ErrorDetail
	for _, e := range errList {
		if detailer, isDetailer := e.(huma.ErrorDetailer); isDetailer {
			details = append(details, detailer.ErrorDetail())
		} else if coded, isCoded := errors.AsType[*errs.Error](e); isCoded && code == "" {
			code = coded.Code
		}
	}
	if code == "" {
		code = statusCode(status)
	}
	if status >= http.StatusInternalServerError {
		msg = errs.PublicMessage(errs.New(code, ""))
	}
	return &errorModel{
		ErrorModel: huma.ErrorModel{
			Status: status,
			Title:  http.StatusText(status),
			Detail: msg,
			Errors: details,
		},
		Code: code,
	}
}

func statusCode(status int) errs.Code {
	switch status {
	case http.StatusUnauthorized:
		return errs.CodeUnauthenticated
	case http.StatusForbidden:
		return errs.CodeForbidden
	case http.StatusNotFound:
		return errs.CodeNotFound
	case http.StatusConflict:
		return errs.CodeConflict
	case http.StatusTooManyRequests:
		return errs.CodeRateLimited
	case http.StatusNotImplemented:
		return errs.CodeNotImplemented
	case http.StatusServiceUnavailable:
		return errs.CodeUnavailable
	}
	if status >= http.StatusInternalServerError {
		return errs.CodeInternal
	}
	return errs.CodeInvalidInput
}

// Error converts err to its response and logs it once, under msg.
func Error(ctx context.Context, msg string, err error) error {
	model := convertStatusError(err)
	recordError(ctx, msg, model, err)
	return model
}

// recordError logs an API error once, at Error for 5xx and Debug otherwise, and adds its code and the
// attributes of its *errs.Error to the request's span.
func recordError(ctx context.Context, msg string, model *errorModel, err error) {
	model.logged = true
	var attrs []attribute.KeyValue
	if coded, isCoded := errors.AsType[*errs.Error](err); isCoded {
		attrs = coded.Attrs
	}
	trace.SpanFromContext(ctx).SetAttributes(append([]attribute.KeyValue{attribute.String("code", string(model.Code))}, attrs...)...)

	logLevel := slog.LevelDebug
	if model.Status >= http.StatusInternalServerError {
		logLevel = slog.LevelError
	}
	logArgs := []any{"status", model.Status, "code", model.Code}
	if err != nil {
		logArgs = append(logArgs, "error", err)
	}
	slog.Log(execution.WithAttrs(ctx, attrs...), logLevel, msg, logArgs...)
}

// convertStatusError maps err to a response status and its public message. Nothing else from err
// reaches the response. An errorModel, such as one from a Huma constructor, is already safe.
func convertStatusError(err error) *errorModel {
	if model, isModel := errors.AsType[*errorModel](err); isModel {
		return model
	}
	if statusErr, isStatusErr := errors.AsType[huma.StatusError](err); isStatusErr {
		return rebuildStatusError(statusErr)
	}

	var details []error
	switch {
	case ent.IsNotFound(err):
		err = errs.Wrap(err, errs.CodeNotFound, "")
	case enumValidationErrFieldRe.MatchString(err.Error()):
		err = errs.Wrap(err, errs.CodeInvalidInput, "")
	case ent.IsConstraintError(err):
		code := errs.CodeInvalidInput
		if match := uniqueErrFieldRe.FindStringSubmatch(err.Error()); match != nil {
			field := match[1]
			if constraintMsg, found := commonConstraints[field]; found {
				details = append(details, NewErrorDetail(constraintMsg, field, nil))
			} else {
				code = errs.CodeConflict
			}
		}
		err = errs.Wrap(err, code, "")
	}

	status, known := codeStatuses[errs.CodeOf(err)]
	if !known {
		status = http.StatusInternalServerError
	}
	return newError(status, errs.PublicMessage(err), append(details, err)...)
}

// rebuildStatusError keeps the status of a status error not built by this package, with its code's
// default text and no details, since those can hold any error's text.
func rebuildStatusError(statusErr huma.StatusError) *errorModel {
	status := statusErr.GetStatus()
	return newError(status, errs.PublicMessage(errs.New(statusCode(status), "")))
}

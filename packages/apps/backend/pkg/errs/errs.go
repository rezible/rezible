// Package errs separates what a caller may see of an error from its cause.
package errs

import (
	"errors"
	"slices"

	"go.opentelemetry.io/otel/attribute"
)

type Code string

const (
	CodeInvalidInput    Code = "invalid_input"
	CodeUnprocessable   Code = "unprocessable"
	CodeUnauthenticated Code = "unauthenticated"
	CodeForbidden       Code = "forbidden"
	CodeNotFound        Code = "not_found"
	CodeConflict        Code = "conflict"
	CodeRateLimited     Code = "rate_limited"
	CodeUnavailable     Code = "unavailable"
	CodeNotImplemented  Code = "not_implemented"
	CodeInternal        Code = "internal"
)

var defaultMessages = map[Code]string{
	CodeInvalidInput:    "Some of the details are invalid.",
	CodeUnprocessable:   "The request can't be completed as sent.",
	CodeUnauthenticated: "Sign in to continue.",
	CodeForbidden:       "You don't have permission to do this.",
	CodeNotFound:        "It doesn't exist or was removed.",
	CodeConflict:        "It conflicts with something that already exists.",
	CodeRateLimited:     "Too many requests. Try again shortly.",
	CodeUnavailable:     "Temporarily unavailable. Try again shortly.",
	CodeNotImplemented:  "Not available yet.",
	CodeInternal:        "Something went wrong.",
}

// Codes returns every code, sorted.
func Codes() []Code {
	codes := make([]Code, 0, len(defaultMessages))
	for code := range defaultMessages {
		codes = append(codes, code)
	}
	slices.Sort(codes)
	return codes
}

// Sentinels for failures callers tell apart with errors.Is.
var (
	ErrTenantContextMissing = Wrap(errors.New("tenant access context not set"), CodeInternal, "")
	ErrForbidden            = New(CodeForbidden, "")
	ErrAuthSessionMissing   = Wrap(errors.New("auth session missing"), CodeUnauthenticated, "")
	ErrAuthSessionExpired   = Wrap(errors.New("auth session expired"), CodeUnauthenticated, "")
	ErrAuthSessionInvalid   = Wrap(errors.New("auth session invalid"), CodeUnauthenticated, "")
	ErrConflict             = New(CodeConflict, "")
	ErrInvalidInput         = New(CodeInvalidInput, "")
	ErrUnprocessableInput   = New(CodeUnprocessable, "")
	ErrNotFound             = New(CodeNotFound, "")
	ErrNotImplemented       = New(CodeNotImplemented, "")
	ErrRateLimited          = New(CodeRateLimited, "")
)

type Error struct {
	Code   Code
	Public string               // safe for a caller to see; empty means the code's default text
	Attrs  []attribute.KeyValue // added to the boundary's log line and span
	Err    error                // the cause; never shown to a caller
}

// Error is for logs: the public text (or the code), then the cause.
func (e *Error) Error() string {
	text := e.Public
	if text == "" {
		text = string(e.Code)
	}
	if e.Err != nil {
		return text + ": " + e.Err.Error()
	}
	return text
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(code Code, public string, attrs ...attribute.KeyValue) *Error {
	return &Error{Code: code, Public: public, Attrs: attrs}
}

func Wrap(err error, code Code, public string, attrs ...attribute.KeyValue) *Error {
	return &Error{Code: code, Public: public, Attrs: attrs, Err: err}
}

// CodeOf returns the code of the outermost *Error in the chain, or CodeInternal if there is none.
func CodeOf(err error) Code {
	if e, ok := errors.AsType[*Error](err); ok {
		return e.Code
	}
	return CodeInternal
}

// PublicMessage returns the public text of the outermost *Error in the chain, or its code's default text.
func PublicMessage(err error) string {
	e, ok := errors.AsType[*Error](err)
	if !ok {
		return defaultMessages[CodeInternal]
	}
	if e.Public != "" {
		return e.Public
	}
	if text, known := defaultMessages[e.Code]; known {
		return text
	}
	return defaultMessages[CodeInternal]
}

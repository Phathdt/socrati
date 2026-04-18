package errors

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

// AppError represents application-level errors with HTTP context
type AppError interface {
	error
	StatusCode() int
	ErrorCode() string
	Message() string
	Details() string
}

// BaseAppError implements AppError interface
type BaseAppError struct {
	Code       string `json:"code"`
	Msg        string `json:"message"`
	Detail     string `json:"details,omitempty"`
	HttpStatus int    `json:"-"`
}

func (e *BaseAppError) Error() string {
	if e.Detail != "" {
		return e.Msg + ": " + e.Detail
	}
	return e.Msg
}

func (e *BaseAppError) StatusCode() int   { return e.HttpStatus }
func (e *BaseAppError) ErrorCode() string { return e.Code }
func (e *BaseAppError) Message() string   { return e.Msg }
func (e *BaseAppError) Details() string   { return e.Detail }

// Wrap returns a new error that preserves the root cause and captures a stack trace.
// Safe to call on sentinel errors — does not mutate the original.
// The stack trace is included in Error() for logging but not in the API response.
func (e *BaseAppError) Wrap(err error) AppError {
	return &WrappedAppError{
		BaseAppError: BaseAppError{
			Code:       e.Code,
			Msg:        e.Msg,
			Detail:     err.Error(),
			HttpStatus: e.HttpStatus,
		},
		rootCause: err,
		stack:     callers(2),
	}
}

// WrappedAppError extends BaseAppError with root cause and stack trace
type WrappedAppError struct {
	BaseAppError
	rootCause error
	stack     string
}

// Error returns the full error with stack trace for logging
func (e *WrappedAppError) Error() string {
	return fmt.Sprintf("%s: %s\n%s", e.Msg, e.rootCause.Error(), e.stack)
}

// Unwrap returns the root cause for errors.Is/As compatibility
func (e *WrappedAppError) Unwrap() error {
	return e.rootCause
}

// callers captures a compact stack trace, skipping n frames
func callers(skip int) string {
	pcs := make([]uintptr, 10)
	n := runtime.Callers(skip+1, pcs)
	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pcs[:n])
	var b strings.Builder
	for {
		frame, more := frames.Next()
		// Skip runtime internals
		if strings.Contains(frame.Function, "runtime.") {
			if !more {
				break
			}
			continue
		}
		fmt.Fprintf(&b, "  at %s (%s:%d)\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	return b.String()
}

// Common reusable errors
var (
	ErrMissingUserID = &BaseAppError{
		Code:       "MISSING_USER_ID",
		Msg:        "x-user-id header is required",
		HttpStatus: http.StatusBadRequest,
	}
	ErrInvalidBody = &BaseAppError{
		Code:       "INVALID_BODY",
		Msg:        "Invalid request body",
		HttpStatus: http.StatusBadRequest,
	}
)

// ErrInvalidBodyWith returns an invalid body error with details
func ErrInvalidBodyWith(details string) AppError {
	return &BaseAppError{
		Code:       "INVALID_BODY",
		Msg:        "Invalid request body",
		Detail:     details,
		HttpStatus: http.StatusBadRequest,
	}
}

// Common error constructors
func NewBadRequestError(code, message, details string) AppError {
	return &BaseAppError{Code: code, Msg: message, Detail: details, HttpStatus: http.StatusBadRequest}
}

func NewNotFoundError(code, message, details string) AppError {
	return &BaseAppError{Code: code, Msg: message, Detail: details, HttpStatus: http.StatusNotFound}
}

func NewConflictError(code, message, details string) AppError {
	return &BaseAppError{Code: code, Msg: message, Detail: details, HttpStatus: http.StatusConflict}
}

func NewInternalError(code, message, details string) AppError {
	return &BaseAppError{Code: code, Msg: message, Detail: details, HttpStatus: http.StatusInternalServerError}
}

func NewValidationError(details string) AppError {
	return &BaseAppError{
		Code: "VALIDATION_ERROR", Msg: "Invalid request format or missing required fields",
		Detail: details, HttpStatus: http.StatusBadRequest,
	}
}

func NewParseError(details string) AppError {
	return &BaseAppError{
		Code:       "PARSE_ERROR",
		Msg:        "Invalid parameter format",
		Detail:     details,
		HttpStatus: http.StatusBadRequest,
	}
}

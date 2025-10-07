package xerrors

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
)

// Error is the interface implemented by extended errors.
type Error interface {
	error
	json.Marshaler

	// Attrs should return a map of attributes associated with the error.
	Attrs() map[string]any

	// Caller should return the information on where the error was generated.
	Caller() CallerInfo

	// Code should return the error code.
	Code() int

	// Is should return true if the wrapped error inside the object matches the given error, false otherwise.
	Is(error) bool

	// String should return a string representation of the error.
	//
	// Unlike the Error() method, this function may include additional information such as the caller details or
	// attributes in any format (eg: plaintext or JSON).
	String() string

	// WithAttr should add an attribute to the error and return itself.
	WithAttr(key string, value any) Error

	// WithAttrs should add attributes to the error and return itself.
	WithAttrs(attrs map[string]any) Error
}

// xerr is a struct that implements the [Error] interface.
type xerr struct {
	// unexported variables
	attrs      map[string]any // error attributes
	caller     *CallerInfo    // information on where the error was generated
	code       int            // the error code
	message    string         // the error message
	wrappedErr error          // the wrapped error, if any
}

// jsonXErr is a version of [xerr] that is used to marshal the object to JSON.
type jsonXErr struct {
	// Attrs is a map of attributes associated with the error.
	Attrs map[string]any `json:"attrs,omitempty"`

	// Caller contains the information on where the error was generated.
	Caller *CallerInfo `json:"caller,omitempty"`

	// Code is the error code.
	Code int `json:"code"`

	// Message is the error message.
	Message string `json:"message"`

	// WrappedError is the wrapped error, if any.
	WrappedError error `json:"wrappedError,omitempty"`
}

// jsonStdErr is a version of a standard Go error that is used to marshal the object to JSON.
type jsonStdError struct {
	// Message is the error message.
	Message string `json:"message"`
}

func (e *jsonStdError) Error() string {
	return e.Message
}

// New creates a new [Error] with the given code and message.
func New(code int, message string) Error {
	xerr := &xerr{
		code:    code,
		message: message,
	}
	if _captureCaller {
		xerr.caller = GetCallerInfo(0)
	}
	return xerr
}

// Newf creates a new [Error] with the given code and formatted message.
func Newf(code int, format string, args ...any) Error {
	xerr := &xerr{
		code:    code,
		message: fmt.Sprintf(format, args...),
	}
	if _captureCaller {
		xerr.caller = GetCallerInfo(0)
	}
	return xerr
}

// Wrap wraps the given error in a new [Error] with the given code and message.
func Wrap(code int, err error, message string) Error {
	xerr := &xerr{
		code:       code,
		message:    message,
		wrappedErr: err,
	}
	if _captureCaller {
		xerr.caller = GetCallerInfo(0)
	}
	return xerr
}

// Wrapf wraps the given error in a new [Error] with the given code and formatted message.
func Wrapf(code int, err error, format string, args ...any) Error {
	xerr := &xerr{
		code:       code,
		message:    fmt.Sprintf(format, args...),
		wrappedErr: err,
	}
	if _captureCaller {
		xerr.caller = GetCallerInfo(0)
	}
	return xerr
}

// Attrs returns a map of attributes associated with the error.
func (e *xerr) Attrs() map[string]any {
	return e.attrs
}

// Caller returns the information on where the error was generated.
func (e *xerr) Caller() CallerInfo {
	if e.caller == nil {
		caller := DefaultCallerInfo()
		return *caller
	}
	return *e.caller
}

// Code returns the error code.
func (e *xerr) Code() int {
	return e.code
}

// Error returns the error message.
func (e *xerr) Error() string {
	return e.message
}

// Is returns true if the error matches the wrapped error in this object (if there is one) or false otherwise.
func (e *xerr) Is(err error) bool {
	if e.wrappedErr == nil {
		return false
	}
	return errors.Is(err, e.wrappedErr)
}

// MarshalJSON marshals the error to JSON.
func (e *xerr) MarshalJSON() ([]byte, error) {
	jsonError := jsonXErr{
		Caller:  e.caller,
		Code:    e.code,
		Message: e.message,
	}
	if e.wrappedErr != nil {
		if _, ok := e.wrappedErr.(Error); !ok {
			jsonError.WrappedError = &jsonStdError{
				Message: e.wrappedErr.Error(),
			}
		}
	}
	if e.attrs != nil {
		jsonError.Attrs = make(map[string]any)
		maps.Copy(jsonError.Attrs, e.attrs)
	}
	return json.Marshal(jsonError)
}

// String returns the error (including the code, attributes, caller and wrapped error) represented as a JSON string.
func (e *xerr) String() string {
	str, err := e.MarshalJSON()
	if err != nil {
		return fmt.Sprintf("failed to marshal error to JSON: %s", err.Error())
	}
	return string(str)
}

// WithAttr adds an attribute to the error and returns itself.
func (e *xerr) WithAttr(key string, value any) Error {
	if e.attrs == nil {
		e.attrs = make(map[string]any)
	}
	e.attrs[key] = value
	return e
}

// WithAttrs adds attributes to the error and returns itself.
func (e *xerr) WithAttrs(attrs map[string]any) Error {
	if e.attrs == nil {
		e.attrs = make(map[string]any)
	}
	maps.Copy(e.attrs, attrs)
	return e
}

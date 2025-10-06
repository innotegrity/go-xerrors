package xerrors

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"runtime"
	"sync"
)

const (
	_unknownFilename = "???"
)

var (
	_captureCaller = false
	_callerMutex   sync.Mutex
)

// CaptureCaller controls whether the caller file and line are captured when a new error is generated.
//
// This function enables or disables the capture of the caller information globally for this package.  This call is
// thread-safe.
func CaptureCaller(enable bool) {
	_callerMutex.Lock()
	_captureCaller = enable
	_callerMutex.Unlock()
}

// getCaller is a helper function to fetch caller information for an error.
func getCaller(skip int) (file string, line int) {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		file = "???"
		line = 0
	}
	return file, line
}

// Error is the interface implemented by extended errors.
type Error interface {
	error
	json.Marshaler

	// Attrs should return a map of attributes associated with the error.
	Attrs() map[string]any

	// Code should return the error code.
	Code() int

	// File should return the file where the error was generated.
	File() string

	// Is should return true if the wrapped error inside the object matches the given error, false otherwise.
	Is(error) bool

	// Line should return the line where the error was generated.
	Line() int

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

// xerr is a struct that implements the Error interface.
type xerr struct {
	// unexported variables
	attrs      map[string]any // error attributes
	code       int            // the error code
	file       string         // the file where the error was generated
	line       int            // the line where the error was generated
	message    string         // the error message
	wrappedErr error          // the wrapped error, if any
}

// jsonXErr is a version of [Err] that is used to marshal an Err object to JSON.
type jsonXErr struct {
	// Attrs is a map of attributes associated with the error.
	Attrs map[string]any `json:"attrs,omitempty"`

	// Code is the error code.
	Code int `json:"code"`

	// File is the file where the error was generated.
	File *string `json:"file,omitempty"`

	// Line is the line where the error was generated.
	Line *int `json:"line,omitempty"`

	// Message is the error message.
	Message string `json:"message"`
}

// New creates a new [Error] with the given code and message.
func New(code int, message string) Error {
	xerr := &xerr{
		code:    code,
		message: message,
	}
	if _captureCaller {
		xerr.file, xerr.line = getCaller(1)
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
		xerr.file, xerr.line = getCaller(1)
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
		xerr.file, xerr.line = getCaller(1)
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
		xerr.file, xerr.line = getCaller(1)
	}
	return xerr
}

// Attrs returns a map of attributes associated with the error.
func (e *xerr) Attrs() map[string]any {
	return e.attrs
}

// Code returns the error code.
func (e *xerr) Code() int {
	return e.code
}

// Error returns the error message.
func (e *xerr) Error() string {
	return e.message
}

// File returns the file where the error was generated.
func (e *xerr) File() string {
	return e.file
}

// Line returns the line where the error was generated.
func (e *xerr) Line() int {
	return e.line
}

// Is returns true if the error matches the wrapped error (if there is one), false otherwise.
func (e *xerr) Is(err error) bool {
	if e.wrappedErr == nil {
		return false
	}
	return errors.Is(err, e.wrappedErr)
}

// MarshalJSON marshals the error to JSON.
func (e *xerr) MarshalJSON() ([]byte, error) {
	jsonError := jsonXErr{
		Code:    e.code,
		Message: e.message,
	}
	if e.attrs != nil {
		maps.Copy(jsonError.Attrs, e.attrs)
	}
	if e.file != _unknownFilename && e.file != "" {
		jsonError.File = &e.file
	}
	if e.line >= 1 {
		jsonError.Line = &e.line
	}
	return json.Marshal(jsonError)
}

// String returns the error (including the code, attributes and any caller information) represented as a JSON string.
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

package xerrors

import (
	"encoding/json"
	"fmt"
	"maps"
	"runtime"
)

const (
	_unknownFilename = "???"
)

var (
	_captureCaller = false
)

// CaptureCaller controls whether the caller file and line are captured when an error is generated.
func CaptureCaller(enable bool) {
	_captureCaller = enable
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

	// Is should return true if the error code matches the given code, false otherwise.
	Is(int) bool

	// Line should return the line where the error was generated.
	Line() int

	// String should return a string representation of the error.
	//
	// Unlike the Error() method, this function may include additional information such as the caller details or
	// attributes in any format (eg: plaintext or JSON).
	String() string
}

// Err is a struct that implements the Error interface.
type Err struct {
	// unexported variables
	attrs   map[string]any // error attributes
	code    int            // the error code
	file    string         // the file where the error was generated
	line    int            // the line where the error was generated
	message string         // the error message
}

// jsonErr is a version of [Err] that is used to marshal an Err object to JSON.
type jsonErr struct {
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

// New creates a new error with the given code and message.
func New(code int, message string) *Err {
	err := &Err{
		code:    code,
		message: message,
	}
	if _captureCaller {
		err.file, err.line = getCaller(1)
	}
	return err
}

// Newf creates a new error with the given code and formatted message.
func Newf(code int, format string, args ...any) *Err {
	err := &Err{
		code:    code,
		message: fmt.Sprintf(format, args...),
	}
	if _captureCaller {
		err.file, err.line = getCaller(1)
	}
	return err
}

// Attrs returns a map of attributes associated with the error.
func (e *Err) Attrs() map[string]any {
	return e.attrs
}

// Code returns the error code.
func (e *Err) Code() int {
	return e.code
}

// Error returns the error message.
func (e *Err) Error() string {
	return e.message
}

// File returns the file where the error was generated.
func (e *Err) File() string {
	return e.file
}

// Line returns the line where the error was generated.
func (e *Err) Line() int {
	return e.line
}

// Is returns true if the error code matches the given code, false otherwise.
func (e *Err) Is(code int) bool {
	return e.code == code
}

// MarshalJSON marshals the error to JSON.
func (e *Err) MarshalJSON() ([]byte, error) {
	jsonError := jsonErr{
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
func (e *Err) String() string {
	str, err := e.MarshalJSON()
	if err != nil {
		return fmt.Sprintf("failed to marshal error to JSON: %s", err.Error())
	}
	return string(str)
}

// With adds an attribute to the error and returns itself.
func (e *Err) With(key string, value any) *Err {
	if e.attrs == nil {
		e.attrs = make(map[string]any)
	}
	e.attrs[key] = value
	return e
}

// WithAttrs adds attributes to the error and returns itself.
func (e *Err) WithAttrs(attrs map[string]any) *Err {
	if e.attrs == nil {
		e.attrs = make(map[string]any)
	}
	maps.Copy(e.attrs, attrs)
	return e
}

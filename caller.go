package xerrors

import (
	"runtime"
	"strings"
	"sync"
)

const (
	_unknownString = "???"
)

var (
	_captureCaller      = false
	_callerFilePrefixes = []string{}
	_callerMutex        sync.Mutex
)

// CaptureCallerInfo controls whether the caller info should be captured when a new error is generated.
//
// This function enables or disables the capture of the caller information globally for this package.  This call is
// thread-safe.
func CaptureCallerInfo(enable bool) {
	_callerMutex.Lock()
	_captureCaller = enable
	_callerMutex.Unlock()
}

// StripCallerFilePrefixes allows you to specify a list of file prefixes that should be stripped from the file path
// when capturing the caller information.
//
// Only the first matching prefix will be stripped from the file path.
//
// This function affects all [CallerInfo] objects generated globally by this package.  This call is thread-safe.
func StripCallerFilePrefixes(prefixes ...string) {
	_callerMutex.Lock()
	_callerFilePrefixes = prefixes
	_callerMutex.Unlock()
}

// CallerInfo holds information about the location from which the error was generated.
type CallerInfo struct {
	// File is the name of the file in which the error occurred, relative to the package root.
	File string `json:"file"`

	// Line is the line number at which the error occurred.
	Line int `json:"line"`

	// Func is the name of the function in which the error occurred.
	Func string `json:"func"`
}

// DefaultCallerInfo returns a default [CallerInfo] that indicates that no caller information was captured.
func DefaultCallerInfo() *CallerInfo {
	return &CallerInfo{
		File: _unknownString,
		Line: 0,
		Func: _unknownString,
	}
}

// GetCallerInfo retrieves the file path, line number, and function name of the caller, formatting the file path to
// be relative to the package directory.
//
// If the caller information is not available, a default [CallerInfo] is returned.
//
// The 'skip' parameter indicates how many stack frames to ascend with 0 being the immediate caller of this function.
//
// This function does not have to be called directly if you are using the [New], [Newf], [Wrap] or [Wrapf] functions
// to generate errors and you have enabled caller capture using [CaptureCallerInfo].
func GetCallerInfo(skip int) *CallerInfo {
	// runtime.Caller returns the program counter (pc), file path, line number, and success status.
	pc, file, line, ok := runtime.Caller(2 + skip)
	if !ok {
		return DefaultCallerInfo()
	}

	// get the full function name
	fn := runtime.FuncForPC(pc).Name()

	// strip the first matching prefix from the file path
	for _, prefix := range _callerFilePrefixes {
		if strings.HasPrefix(file, prefix) {
			file = file[len(prefix):]
			break
		}
	}

	return &CallerInfo{
		File: file,
		Line: line,
		Func: fn,
	}
}

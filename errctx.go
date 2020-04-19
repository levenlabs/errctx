// Package errctx allows for setting and retrieving contextual information on
// error objects
package errctx

import (
	"fmt"
	"path"
	"runtime"
)

type sourceKey int

type errctx struct {
	err error
	ctx map[interface{}]interface{}
}

// Error implements the error interface
func (ec errctx) Error() string {
	return ec.err.Error()
}

// Unwrap returns the wrapped error. Necessary for errors.Is to work
func (ec errctx) Unwrap() error {
	return ec.err
}

// Is returns true if the sent error has the same underlying error or is the
// underlying error. Without this then errors.Is(err, err2) won't work if they're
// both wrapped errors.
func (ec errctx) Is(err error) bool {
	return err == ec.err || Base(err) == ec.err
}

// Base returns the underlying error object that was prevoiusly wrapped in a
// call to Set. If the error did not come from Set it is returned as-is.
func Base(err error) error {
	if ec, ok := err.(errctx); ok {
		return ec.err
	}
	return err
}

// Set takes in an error and one or more key/value pairs. It returns an error
// instance which can have Get called on it with one of those passed in keys to
// retrieve the associated value later.
//
// Errors returned from Set are immutable. For example:
//
//	err := errors.New("ERR")
//	fmt.Println(errctx.Get(err, "foo")) // ""
//
//	err2 := errctx.Set(err, "foo", "a")
//	fmt.Println(errctx.Get(err2, "foo")) // "a"
//
//	err3 := errctx.Set(err2, "foo", "b")
//	fmt.Println(errctx.Get(err2, "foo")) // "a"
//	fmt.Println(errctx.Get(err3, "foo")) // "b"
//
func Set(err error, kvs ...interface{}) error {
	ec := errctx{
		err: Base(err),
		ctx: map[interface{}]interface{}{},
	}

	if ecinner, ok := err.(errctx); ok {
		for k, v := range ecinner.ctx {
			ec.ctx[k] = v
		}
	}
	for i := 0; i < len(kvs); i += 2 {
		ec.ctx[kvs[i]] = kvs[i+1]
	}
	return ec
}

// Get retrieves the value associated with the key by a previous call to Set,
// which this error should have been returned from. Returns nil if the key isn't
// set, or if the error wasn't previously wrapped by Set at all.
func Get(err error, k interface{}) interface{} {
	ec, ok := err.(errctx)
	if !ok {
		return nil
	}
	return ec.ctx[k]
}

// Mark records the filename and line number that called Mark and sets it on
// the error. Future calls to Mark will NOT overwrite the previous line.
func Mark(err error) error {
	return MarkSkip(err, 1)
}

// MarkSkip is like Mark but allows you to skip an arbitrary amount of
// functions from the stack. Sending skip of 0 means to Mark the caller of this
// function.
func MarkSkip(err error, skip int) error {
	if err == nil {
		return nil
	}
	// check if it was already marked
	if Get(err, sourceKey(0)) != nil {
		return err
	}
	// since 0 means the caller of Caller, 1 means the caller of MarkSkip
	_, file, line, ok := runtime.Caller(1 + skip)
	if !ok {
		return err
	}
	file = path.Base(file)
	return Set(err, sourceKey(0), fmt.Sprintf("%s:%d", file, line))
}

// Line returns the file and line number where Mark was first called on the
// error and a boolean indicating if any line was found.
func Line(err error) (string, bool) {
	ec, ok := err.(errctx)
	if !ok {
		return "", false
	}
	s, ok := ec.ctx[sourceKey(0)].(string)
	return s, ok
}

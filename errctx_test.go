package errctx

import (
	"errors"
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type wrapErr struct {
	err error
}

func (e wrapErr) Error() string {
	return e.err.Error()
}

func (e wrapErr) Unwrap() error {
	return e.err
}

type wrapErrs struct {
	errs []error
}

func (e wrapErrs) Error() string {
	return fmt.Sprintf("%v", e.errs)
}

func (e wrapErrs) Unwrap() []error {
	return e.errs
}

type key int

func TestErrCtx(t *testing.T) {
	err := errors.New("foo")

	assert.Equal(t, err, Base(err))

	err1 := Set(err, key(0), "a")
	assert.Equal(t, err.Error(), err1.Error())
	assert.Equal(t, err, Base(err1))
	assert.True(t, errors.Is(err1, err))
	assert.Nil(t, Get(err, key(0)))
	assert.Equal(t, "a", Get(err1, key(0)))

	err2 := Set(err, key(1), "b")
	assert.NotEqual(t, err1, err2)
	assert.Equal(t, err.Error(), err2.Error())
	assert.Equal(t, err, Base(err2))
	assert.True(t, errors.Is(err2, err))
	assert.Nil(t, Get(err, key(1)))
	assert.Nil(t, Get(err2, key(0)))
	assert.Equal(t, "b", Get(err2, key(1)))

	err3 := Set(err2, key(2), "c")
	assert.Equal(t, err.Error(), err3.Error())
	assert.Equal(t, err, Base(err3))
	assert.True(t, errors.Is(err3, err))
	assert.Nil(t, Get(err3, key(0)))
	assert.Nil(t, Get(err2, key(2)))
	assert.Equal(t, "b", Get(err3, key(1)))
	assert.Equal(t, "c", Get(err3, key(2)))

	assert.True(t, err3.(errctx).Is(err3))
	assert.True(t, err3.(errctx).Is(err2))
	assert.True(t, err3.(errctx).Is(err2.(errctx).err))

	err4 := wrapErr{err: err3}
	assert.Equal(t, err4, Base(err4))
	assert.Nil(t, Get(err4, key(0)))
	assert.Equal(t, "b", Get(err4, key(1)))

	err5 := Set(err4, key(3), "d")
	assert.Equal(t, err4, Base(err5))
	assert.Nil(t, Get(err5, key(0)))
	assert.Equal(t, "b", Get(err5, key(1)))
	assert.Equal(t, "c", Get(err5, key(2)))
	assert.Equal(t, "d", Get(err5, key(3)))

	err6 := wrapErrs{errs: []error{wrapErr{err: err5}, err1}}
	assert.Equal(t, err6, Base(err6))
	assert.Equal(t, "a", Get(err6, key(0)))
	assert.Equal(t, "b", Get(err6, key(1)))
	assert.Equal(t, "c", Get(err6, key(2)))
	assert.Equal(t, "d", Get(err6, key(3)))

	err7 := Set(err6, key(4), "e")
	assert.Equal(t, err6, Base(err7))
	assert.Equal(t, "a", Get(err7, key(0)))
	assert.Equal(t, "b", Get(err7, key(1)))
	assert.Equal(t, "c", Get(err7, key(2)))
	assert.Equal(t, "d", Get(err7, key(3)))
	assert.Equal(t, "e", Get(err7, key(4)))

	assert.Nil(t, Get(wrapErr{}, key(0)))
	assert.Nil(t, Get(wrapErrs{}, key(0)))
	assert.Nil(t, Get(wrapErrs{errs: []error{nil}}, key(0)))
	assert.Equal(t, "a", Get(wrapErrs{errs: []error{nil, err1}}, key(0)))
	assert.Nil(t, Get(nil, key(0)))
}

func TestMark(t *testing.T) {
	err := errors.New("bar")

	l, ok := Line(err)
	assert.False(t, ok)
	assert.Empty(t, l)

	_, _, ln, ok := runtime.Caller(0)
	require.True(t, ok)
	err = Mark(err)
	l, ok = Line(err)
	assert.True(t, ok)
	assert.Equal(t, fmt.Sprintf("errctx_test.go:%d", ln+2), l)
	l, ok = Line(wrapErr{err: err})
	assert.True(t, ok)
	assert.Equal(t, fmt.Sprintf("errctx_test.go:%d", ln+2), l)

	// calling it again shouldn't do anything
	err = Mark(err)
	l, ok = Line(err)
	assert.True(t, ok)
	assert.Equal(t, fmt.Sprintf("errctx_test.go:%d", ln+2), l)

	err = func() error {
		// 1 should return the anonymous function
		return MarkSkip(errors.New("bar"), 1)
	}()
	_, _, ln, ok = runtime.Caller(0)
	require.True(t, ok)
	l, ok = Line(err)
	assert.True(t, ok)
	assert.Equal(t, fmt.Sprintf("errctx_test.go:%d", ln-1), l)

	assert.Nil(t, Mark(nil))
	assert.Nil(t, MarkSkip(nil, 1))
	_, ok = Line(nil)
	assert.False(t, ok)
}

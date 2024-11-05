package check

import (
	"encoding/json"
	"io/fs"
	"os"
	"syscall"
	"testing"
)

func testCheck(msg string, ok bool) func(t *testing.T, expect bool) {
	return func(t *testing.T, expect bool) {
		if !expect && msg == "" {
			t.Helper()
			t.Error("expected a fail message, got nothing")
		}

		if expect == ok {
			return
		}

		t.Helper()
		if !ok {
			t.Errorf("test failed, expected success")
		} else {
			t.Errorf("test succeeded, expected failure")
		}
	}
}

func TestCheckTrue(t *testing.T) {
	testCheck(checkTrue(true))(t, true)
	testCheck(checkTrue(false))(t, false)
}

func TestCheckFalse(t *testing.T) {
	testCheck(checkFalse(false))(t, true)
	testCheck(checkFalse(true))(t, false)
}

func TestCheckEqual(t *testing.T) {
	testCheck(checkEqual(1, 1))(t, true)
	testCheck(checkEqual(1, 2))(t, false)
	testCheck(checkEqual(int8(1), int16(1)))(t, false)
}

func TestCheckNotEqual(t *testing.T) {
	testCheck(checkNotEqual(1, 2))(t, true)
	testCheck(checkNotEqual(1, 1))(t, false)
	testCheck(checkNotEqual(int8(1), int16(2)))(t, false)
}

func TestCheckNil(t *testing.T) {
	testCheck(checkNil(nil))(t, true)
	testCheck(checkNil(new(int)))(t, false)
}

func TestCheckNotNil(t *testing.T) {
	testCheck(checkNotNil(new(int)))(t, true)
	testCheck(checkNotNil(nil))(t, false)
}

func TestCheckZero(t *testing.T) {
	testCheck(checkZero(nil))(t, true)
	testCheck(checkZero(0))(t, true)
	testCheck(checkZero((*int)(nil)))(t, true)
	testCheck(checkZero(1))(t, false)
	testCheck(checkZero(new(int)))(t, false)
}

func TestCheckNotZero(t *testing.T) {
	testCheck(checkNotZero(1))(t, true)
	testCheck(checkNotZero(new(int)))(t, true)
	testCheck(checkNotZero(nil))(t, false)
	testCheck(checkNotZero(0))(t, false)
	testCheck(checkNotZero((*int)(nil)))(t, false)
}

func TestCheckErrIs(t *testing.T) {
	err := &os.PathError{
		Op:   "test",
		Path: "test",
		Err:  syscall.EPERM,
	}

	testCheck(checkErrIs(err, fs.ErrPermission))(t, true)
	testCheck(checkErrIs(err, fs.ErrClosed))(t, false)
}

func TestCheckErrAs(t *testing.T) {
	var (
		goodTarget syscall.Errno
		badTarget  *json.InvalidUnmarshalError
		err        = &os.PathError{
			Op:   "test",
			Path: "test",
			Err:  syscall.EPERM,
		}
	)

	testCheck(checkErrAs(err, &goodTarget))(t, true)
	testCheck(checkErrAs(err, &badTarget))(t, false)
}

func TestCheckHasKey(t *testing.T) {
	m := map[string]string{
		"k": "v",
	}

	testCheck(checkHasKey(m, "k"))(t, true)
	testCheck(checkHasKey(m, "v"))(t, false)
	testCheck(checkHasKey(m, 1))(t, false)

	testCheck(checkHasKey(1, 2))(t, false)
}

func TestCheckNotHasKey(t *testing.T) {
	m := map[string]string{
		"k": "v",
	}

	testCheck(checkNotHasKey(m, "v"))(t, true)
	testCheck(checkNotHasKey(m, "k"))(t, false)
	testCheck(checkNotHasKey(m, 1))(t, false)

	testCheck(checkNotHasKey(1, 2))(t, false)
}

func TestCheckContains(t *testing.T) {
	t.Run("Map", func(t *testing.T) {
		m := map[string]string{
			"k": "v",
		}

		testCheck(checkContains(m, "v"))(t, true)
		testCheck(checkContains(m, "k"))(t, false)
	})

	t.Run("Slice", func(t *testing.T) {
		s := []string{"a"}

		testCheck(checkContains(s, "a"))(t, true)
		testCheck(checkContains(s, "b"))(t, false)
	})

	t.Run("Array", func(t *testing.T) {
		a := [...]string{"a"}

		testCheck(checkContains(a, "a"))(t, true)
		testCheck(checkContains(a, "b"))(t, false)
	})

	t.Run("String", func(t *testing.T) {
		s := "hello world"

		testCheck(checkContains(s, "hello"))(t, true)
		testCheck(checkContains(s, "goodbye"))(t, false)

		testCheck(checkContains("test", 123))(t, false)
	})

	testCheck(checkContains(123, 1))(t, false)
}

func TestCheckNotContains(t *testing.T) {
	t.Run("Map", func(t *testing.T) {
		m := map[string]string{
			"k": "v",
		}

		testCheck(checkNotContains(m, "k"))(t, true)
		testCheck(checkNotContains(m, "v"))(t, false)
	})

	t.Run("Slice", func(t *testing.T) {
		s := []string{"a"}

		testCheck(checkNotContains(s, "b"))(t, true)
		testCheck(checkNotContains(s, "a"))(t, false)
	})

	t.Run("Array", func(t *testing.T) {
		a := [...]string{"a"}

		testCheck(checkNotContains(a, "b"))(t, true)
		testCheck(checkNotContains(a, "a"))(t, false)
	})

	t.Run("String", func(t *testing.T) {
		s := "hello world"

		testCheck(checkNotContains(s, "goodbye"))(t, true)
		testCheck(checkNotContains(s, "hello"))(t, false)

		testCheck(checkNotContains("test", 123))(t, false)
	})

	testCheck(checkNotContains(123, 1))(t, false)
}

func TestCheckPanics(t *testing.T) {
	testCheck(checkPanics(func() { panic("check") }))(t, true)
	testCheck(checkPanics(func() {}))(t, false)
}

func TestCheckNotPanics(t *testing.T) {
	testCheck(checkNotPanics(func() {}))(t, true)
	testCheck(checkNotPanics(func() { panic("check") }))(t, false)
}

func TestCheckPanicsWith(t *testing.T) {
	r := 0xb374

	testCheck(checkPanicsWith(r, func() { panic(r) }))(t, true)
	testCheck(checkPanicsWith(r, func() { panic(0) }))(t, false)
	testCheck(checkPanicsWith(r, func() {}))(t, false)
}

func TestCheckEventuallyTrue(t *testing.T) {
	testCheck(checkEventuallyTrue(100, func(i int) bool { return true }))(t, true)
	testCheck(checkEventuallyTrue(100, func(i int) bool { return false }))(t, false)
}

func TestCheckEventuallyNil(t *testing.T) {
	testCheck(checkEventuallyNil(100, func(i int) error { return nil }))(t, true)
	testCheck(checkEventuallyNil(100, func(i int) error { return os.ErrClosed }))(t, false)
}

package check

import (
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"unicode"

	"github.com/peter-evans/patience"
	"github.com/thatguystone/cog/textwrap"
)

type Error interface {
	Helper()
	Error(args ...any)
}

type Fatal interface {
	Helper()
	Fatal(args ...any)
}

func checkTrue(cond bool) (string, bool) {
	if cond {
		return "", true
	}

	return "Expected true", false
}

func checkFalse(cond bool) (string, bool) {
	if !cond {
		return "", true
	}

	return "Expected false", false
}

func equalMsg(g, e any) string {
	var (
		gs = dump(g, 0)
		gl = strings.Split(gs, "\n")
		es = dump(e, 0)
		el = strings.Split(es, "\n")
	)

	if len(gl) == 1 && len(el) == 1 {
		return fmt.Sprintf(""+
			"Expected: %s\n"+
			"       == %s",
			gs,
			es,
		)
	}

	var (
		b     = new(strings.Builder)
		diffs = patience.Diff(gl, el)
	)

	const (
		prefixLen = 2
		prelude   = "Expected values to be equal:\n"
	)

	n := len(prelude) + (len(dumpIndent)+prefixLen+1)*len(diffs)
	for _, diff := range diffs {
		n += len(diff.Text)
	}

	b.Grow(n)
	b.WriteString(prelude)

	for i, diff := range diffs {
		if i > 0 {
			b.WriteByte('\n')
		}

		b.WriteString(dumpIndent)

		switch diff.Type {
		case patience.Delete:
			b.WriteString("- ")
		case patience.Insert:
			b.WriteString("+ ")
		default:
			b.WriteString("  ")
		}

		b.WriteString(diff.Text)
	}

	return b.String()
}

func checkEqual(g, e any) (string, bool) {
	if reflect.DeepEqual(g, e) {
		return "", true
	}

	return equalMsg(g, e), false
}

func checkNotEqual(g, e any) (string, bool) {
	if reflect.DeepEqual(g, e) {
		return "Expected values to differ:\n" + dump(g, 1), false
	}

	// Not equal because of a type mismatch is probably not what was meant
	if reflect.TypeOf(g) != reflect.TypeOf(e) {
		msg := fmt.Sprintf("Cannot compare mismatched types: %T != %T", g, e)
		return msg, false
	}

	return "", true
}

func checkNil(v any) (string, bool) {
	if v == nil {
		return "", true
	}

	return "Expected nil, got:\n" + dump(v, 1), false
}

func checkNotNil(v any) (string, bool) {
	if v != nil {
		return "", true
	}

	return "Expected something, got nil", false
}

func checkZero(v any) (string, bool) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.IsZero() {
		return "", true
	}

	return equalMsg(v, reflect.Zero(rv.Type()).Interface()), false
}

func checkNotZero(v any) (string, bool) {
	rv := reflect.ValueOf(v)
	if rv.IsValid() && !rv.IsZero() {
		return "", true
	}

	return "Expected something, got {}", false
}

func checkErrIs(err, target error) (string, bool) {
	if errors.Is(err, target) {
		return "", true
	}

	return equalMsg(err, target), false
}

func checkErrAs(err error, target any) (string, bool) {
	if errors.As(err, target) {
		return "", true
	}

	return equalMsg(err, target), false
}

func containsMsg(container any, what string, el any) string {
	return "Expected to find " + what + " in iter:\n" +
		dumpIndent + upperFirst(what) + ":\n" +
		dump(el, 2) +
		"\n" +
		dumpIndent + "Iter:\n" +
		dump(container, 2)
}

func notContainsMsg(container any, what string, el any) string {
	return "Unexpectedly found " + what + " in iter:\n" +
		dumpIndent + upperFirst(what) + ":\n" +
		dump(el, 2) +
		"\n" +
		dumpIndent + "Iter:\n" +
		dump(container, 2)
}

func upperFirst(str string) string {
	first := true

	return strings.Map(
		func(r rune) rune {
			if first {
				r = unicode.ToTitle(r)
				first = false
			}

			return r
		},
		str)
}

func hasKey(m, k any) (string, bool) {
	mv := reflect.ValueOf(m)
	if mv.Kind() != reflect.Map {
		msg := fmt.Sprintf("Cannot check non-map %T for key", m)
		return msg, false
	}

	kv := reflect.ValueOf(k)
	if !kv.Type().AssignableTo(mv.Type().Key()) {
		msg := fmt.Sprintf("Cannot check %T for mismatched key type %T", m, k)
		return msg, false
	}

	return "", mv.MapIndex(kv).IsValid()
}

func checkHasKey(m, k any) (string, bool) {
	msg, ok := hasKey(m, k)
	if msg != "" {
		return msg, false
	}

	if ok {
		return "", true
	}

	return containsMsg(m, "key", k), false
}

func checkNotHasKey(m, k any) (string, bool) {
	msg, ok := hasKey(m, k)
	if msg != "" {
		return msg, false
	}

	if !ok {
		return "", true
	}

	return notContainsMsg(m, "key", k), false
}

func contains(iter, v any) (msg, what string, ok bool) {
	rv := reflect.ValueOf(iter)

	switch rv.Kind() {
	case reflect.Map:
		what = "value"

		for mi := rv.MapRange(); mi.Next(); {
			if reflect.DeepEqual(mi.Value().Interface(), v) {
				ok = true
				return
			}
		}

		return

	case reflect.Slice, reflect.Array:
		what = "value"

		for i := range rv.Len() {
			if reflect.DeepEqual(rv.Index(i).Interface(), v) {
				ok = true
				return
			}
		}

		return

	case reflect.String:
		what = "substring"

		substr, isStr := v.(string)
		if !isStr {
			msg = fmt.Sprintf("Cannot search string for non-string %T", v)
			return
		}

		if strings.Contains(rv.String(), substr) {
			ok = true
			return
		}

		return

	default:
		msg = fmt.Sprintf("Cannot check non-container %T for containment", iter)
		return
	}
}

func checkContains(iter, v any) (string, bool) {
	msg, what, ok := contains(iter, v)
	if msg != "" {
		return msg, false
	}

	if ok {
		return "", true
	}

	return containsMsg(iter, what, v), false
}

func checkNotContains(iter, v any) (string, bool) {
	msg, what, ok := contains(iter, v)
	if msg != "" {
		return msg, false
	}

	if !ok {
		return "", true
	}

	return notContainsMsg(iter, what, v), false
}

func checkPanics(fn func()) (msg string, ok bool) {
	defer func() {
		ok = recover() != nil
	}()

	fn()
	return "Expected func to panic", false
}

func checkNotPanics(fn func()) (msg string, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			msg = "Expected func not to panic:\n" +
				dump(r, 1) +
				"\n" +
				"\n" +
				textwrap.Indent(string(debug.Stack()), dumpIndent)
		}
	}()

	fn()
	return "", true
}

func checkPanicsWith(recovers any, fn func()) (msg string, ok bool) {
	defer func() {
		r := recover()
		if r == nil {
			msg = "Expected func to panic"
			return
		}

		if !reflect.DeepEqual(r, recovers) {
			msg = equalMsg(r, recovers) +
				"\n" +
				"\n" +
				textwrap.Indent(string(debug.Stack()), dumpIndent)
			return
		}

		ok = true
	}()

	fn()
	return
}

func checkEventuallyTrue(numTries int, fn func(i int) bool) (string, bool) {
	for i := range numTries {
		if fn(i) {
			return "", true
		}
	}

	msg := fmt.Sprintf(
		"Polling for condition failed, gave up after %d tries",
		numTries,
	)
	return msg, false
}

func checkEventuallyNil(numTries int, fn func(i int) error) (string, bool) {
	var err error

	for i := range numTries {
		err = fn(i)
		if err == nil {
			return "", true
		}
	}

	msg := fmt.Sprintf(
		"Func didn't succeed after %d tries, last err:\n%s",
		numTries,
		dump(err, 1),
	)
	return msg, false
}

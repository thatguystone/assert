//go:build generate

//go:generate go run $GOFILE

package main

import (
	"text/template"

	"github.com/thatguystone/cog/assert"
	"github.com/thatguystone/cog/generate"
)

func newTemplate(tmpl string) *template.Template {
	return assert.Must(template.New("assert").Parse(tmpl))
}

func main() {
	tmpls := []*template.Template{
		newTemplate(`
			// {{ .Doc }}
			func {{ .Name }}(t Error, {{ .Args }}) bool {
				if msg, ok := {{ .Check }}; !ok {
					t.Helper()
					t.Error("\n" + msg)
					return false
				}

				return true
			}
		`),
		newTemplate(`
			// {{ .Doc }}
			func {{ .Name }}f(t Error, {{ .Args }}, format string, args ...any) bool {
				if msg, ok := {{ .Check }}; !ok {
					t.Helper()
					t.Error(fmt.Sprintf(format, args...) + "\n" + msg)
					return false
				}

				return true
			}
		`),
		newTemplate(`
			// {{ .Doc }}
			func Must{{ or .Must .Name  }}(t Fatal, {{ .Args }}) {
				if msg, ok := {{ .Check }}; !ok {
					t.Helper()
					t.Fatal("\n" + msg)
				}
			}
		`),
		newTemplate(`
			// {{ .Doc }}
			func Must{{ or .Must .Name }}f(t Fatal, {{ .Args }}, format string, args ...any) {
				if msg, ok := {{ .Check }}; !ok {
					t.Helper()
					t.Fatal(fmt.Sprintf(format, args...) + "\n" + msg)
				}
			}
		`),
	}

	b := generate.New()
	b.WriteString("\n//gocovr:skip-file\n")
	b.WriteString("\nimport()\n")

	for _, fn := range funcs {
		for _, tmpl := range tmpls {
			err := tmpl.Execute(b, fn)
			assert.Nil(err)
		}
	}

	b.WriteFile()
}

type Func struct {
	Name  string
	Must  string
	Args  string
	Check string
	Doc   string
}

var funcs = []Func{
	{
		Name:  "True",
		Args:  "cond bool",
		Check: "checkTrue(cond)",
		Doc:   "Check that the given bool is true.",
	},
	{
		Name:  "False",
		Args:  "cond bool",
		Check: "checkFalse(cond)",
		Doc:   "Check that the given bool is false.",
	},
	{
		Name:  "Equal",
		Args:  "g, e any",
		Check: "checkEqual(g, e)",
		Doc:   "Check that two things are equal; e is the expected value, g is what was got.",
	},
	{
		Name:  "NotEqual",
		Args:  "g, e any",
		Check: "checkNotEqual(g, e)",
		Doc:   "Check that two things are not equal; e is the expected value, g is what was got.",
	},
	{
		Name:  "Nil",
		Args:  "v any",
		Check: "checkNil(v)",
		Doc:   "Check that v is nil. This is a strict equality check.",
	},
	{
		Name:  "NotNil",
		Args:  "v any",
		Check: "checkNotNil(v)",
		Doc:   "Check that v is not nil. This is a strict equality check.",
	},
	{
		Name:  "Zero",
		Args:  "v any",
		Check: "checkZero(v)",
		Doc:   "Check that v is the zero value for its type.",
	},
	{
		Name:  "NotZero",
		Args:  "v any",
		Check: "checkNotZero(v)",
		Doc:   "Check that v is not the zero value for its type.",
	},
	{
		Name:  "ErrIs",
		Args:  "err, target error",
		Check: "checkErrIs(err, target)",
		Doc:   "Check that [errors.Is] returns true.",
	},
	{
		Name:  "ErrAs",
		Args:  "err error, target any",
		Check: "checkErrAs(err, target)",
		Doc:   "Check that [errors.As] returns true.",
	},
	{
		Name:  "HasKey",
		Must:  "HaveKey",
		Args:  "m, k any",
		Check: "checkHasKey(m, k)",
		Doc:   "Check that map m contains key k.",
	},
	{
		Name:  "NotHasKey",
		Must:  "NotHaveKey",
		Args:  "m, k any",
		Check: "checkNotHasKey(m, k)",
		Doc:   "Check that map m does not contain key k.",
	},
	{
		Name:  "Contains",
		Must:  "Contain",
		Args:  "iter, v any",
		Check: "checkContains(iter, v)",
		Doc:   "Check that iter contains value v. Iter must be one of: map, slice, array, or string.",
	},
	{
		Name:  "NotContains",
		Must:  "NotContain",
		Args:  "iter, v any",
		Check: "checkNotContains(iter, v)",
		Doc:   "Check that iter does not contain value v. Iter must be one of: map, slice, array, or string",
	},
	{
		Name:  "Panics",
		Must:  "Panic",
		Args:  "fn func()",
		Check: "checkPanics(fn)",
		Doc:   "Check that the given function panics.",
	},
	{
		Name:  "NotPanics",
		Must:  "NotPanic",
		Args:  "fn func()",
		Check: "checkNotPanics(fn)",
		Doc:   "Check that the given function does not panic.",
	},
	{
		Name:  "PanicsWith",
		Must:  "PanicWith",
		Args:  "recovers any, fn func()",
		Check: "checkPanicsWith(recovers, fn)",
		Doc:   "Check that the given function panics with the given value.",
	},
	{
		Name:  "EventuallyTrue",
		Args:  "numTries int, fn func(i int) bool",
		Check: "checkEventuallyTrue(numTries, fn)",
		Doc:   "Poll the given function, a max of numTries times, until it returns true.",
	},
	{
		Name:  "EventuallyNil",
		Args:  "numTries int, fn func(i int) error",
		Check: "checkEventuallyNil(numTries, fn)",
		Doc:   "Poll the given function, a max of numTries times, until it doesn't return an error. This is mainly a helper used to exhaust error pathways.",
	},
}

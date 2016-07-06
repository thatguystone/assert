package stringc_test

import (
	"testing"

	"github.com/thatguystone/cog/check"
	"github.com/thatguystone/cog/stringc"
)

func TestTrimLinesBasic(t *testing.T) {
	c := check.New(t)

	tests := []struct {
		in  string
		out string
	}{
		{
			in:  "\n\ntests\n\n",
			out: "tests",
		},
		{
			in:  "    \n\ntests\n   \n   \n",
			out: "tests",
		},
		{
			in:  "  \nsome\nstuff\n",
			out: "some\nstuff",
		},
		{
			in:  "  \n  some  \nstuff  \n",
			out: "  some  \nstuff  ",
		},
		{
			in:  "  \n  \n  ",
			out: "",
		},
	}

	for i, test := range tests {
		out := stringc.TrimLines(test.in)
		c.Equal(test.out, out, "failed at %d", i)
	}
}

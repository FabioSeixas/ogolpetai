package main

import (
	"bytes"
	"flag"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

type testEnv struct {
	args           string
	stdout, stderr bytes.Buffer
}

func (e *testEnv) run() error {
	s := flag.NewFlagSet("ogolpetai_test", flag.ContinueOnError)
	s.SetOutput(&e.stderr)
	return run(s, strings.Fields(e.args), &e.stdout)
}

func TestRun(t *testing.T) {
	happy := map[string]struct{ input, expected string }{
		"urlOnly": {
			input:    "http://abc",
			expected: fmt.Sprintf(`Making 100 requests to "http://abc" with concurrency level %d`, runtime.NumCPU()),
		},
		"allParams": {
			input:    "-c=10 -n=50 http://abc",
			expected: `Making 50 requests to "http://abc" with concurrency level 10`,
		},
	}

	for testName, test := range happy {

		env := &testEnv{args: test.input}

		if err := env.run(); err != nil {
			t.Fatalf("\n[%s] got error: %q", testName, err)
		}
		if out := env.stdout.String(); !strings.Contains(out, test.expected) {
			t.Errorf("\n[%s] \ngot: \n\n%q\n\n expect: %q", testName, out, test.expected)
		}

	}
}

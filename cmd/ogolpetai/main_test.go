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
		"should work if only url param is present": {
			input:    "http://abc",
			expected: fmt.Sprintf(`Making 100 requests to "http://abc" with concurrency level %d`, runtime.NumCPU()),
		},
		"should work if all params are present": {
			input:    "-c=10 -n=50 http://abc",
			expected: `Making 50 requests to "http://abc" with concurrency level 10`,
		},
		"should work if url and -c param are present": {
			input:    "-c=10 http://abc",
			expected: `Making 100 requests to "http://abc" with concurrency level 10`,
		},
		"should work if url and -n param are present": {
			input:    "-n=50 http://abc",
			expected: fmt.Sprintf(`Making 50 requests to "http://abc" with concurrency level %d`, runtime.NumCPU()),
		},
	}

	for testName, test := range happy {

		t.Run(testName, func(t *testing.T) {

			env := &testEnv{args: test.input}

			if err := env.run(); err != nil {
				t.Fatalf("\n[%s] got error: %q", testName, err)
			}
			if out := env.stdout.String(); !strings.Contains(out, test.expected) {
				t.Errorf("\n[%s] \ngot: \n\n%s\n\n expect: %q", testName, out, test.expected)
			}
		})

	}

	sad := map[string]string{
		"should stop if url is missing":         "",
		"should stop if url is invalid":         "://foo",
		"should stop if url scheme is not http": "https://foo",
		"should stop if url host is missing":    "http://",
		"should stop if -c is not a number":     "-c=x http://foo",
		"should stop if -c is zero":             "-c=0 http://foo",
		"should stop if -n is zero":             "-n=0 http://foo",
		"should stop if -n is not a number":     "-n=x http://foo",
		"should stop if -n is negative":         "-n=-1 http://foo",
		"should stop if -c is negative":         "-c=-1 http://foo",
		"should stop if -c is greater than -n":  "-c=2 -n=1 http://foo",
	}

	for testName, input := range sad {

		t.Run(testName, func(t *testing.T) {

			env := &testEnv{args: input}

			if err := env.run(); err == nil {
				t.Fatalf("\n[%s] got nil error", testName)
			}
			if env.stderr.Len() == 0 {
				t.Fatalf("\n[%s]\nstderr = 0 bytes\nexpect > 0", testName)
			}
		})

	}
}

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
	headerArgs     []string
	stdout, stderr bytes.Buffer
}

func (e *testEnv) run() error {
	s := flag.NewFlagSet("ogolpetai_test", flag.ContinueOnError)
	s.SetOutput(&e.stderr)

	_args := append(e.headerArgs, strings.Fields(e.args)...)

	return run(s, _args, &e.stdout)
}

func TestRun(t *testing.T) {
	happy := map[string]struct {
		input, expected string
		headers         []string
	}{
		"should work if only url param is present": {
			input:    "http://abc",
			expected: fmt.Sprintf(`Making 100 GET requests to "http://abc" with concurrency level %d (Timeout=10s)`, runtime.NumCPU()),
		},
		"should work if all params are present": {
			input:    "-c=10 -n=50 http://abc",
			expected: `Making 50 GET requests to "http://abc" with concurrency level 10 (Timeout=10s)`,
		},
		"should work if url and -c params are present": {
			input:    "-c=10 http://abc",
			expected: `Making 100 GET requests to "http://abc" with concurrency level 10 (Timeout=10s)`,
		},
		"should work if url and -n params are present": {
			input:    "-n=50 http://abc",
			expected: fmt.Sprintf(`Making 50 GET requests to "http://abc" with concurrency level %d (Timeout=10s)`, runtime.NumCPU()),
		},
		"should work if url and -t params are present": {
			input:    "-t=5s http://abc",
			expected: fmt.Sprintf(`Making 100 GET requests to "http://abc" with concurrency level %d (Timeout=5s)`, runtime.NumCPU()),
		},
		"should work if url and -m params are present": {
			input:    "-m=POST http://abc",
			expected: fmt.Sprintf(`Making 100 POST requests to "http://abc" with concurrency level %d (Timeout=10s)`, runtime.NumCPU()),
		},
		"should work if url and -H params are present": {
			input:    "http://abc",
			headers:  []string{"-H='Content-Type: application/json'"},
			expected: `Headers: 'Content-Type: application/json'`,
		},
		"should work for more than one -H param": {
			input:    "http://abc",
			headers:  []string{"-H='Content-Type: application/json'", "-H='User-Agent: ogolpetai'"},
			expected: `Headers: 'Content-Type: application/json', 'User-Agent: ogolpetai'`,
		},
	}

	for testName, test := range happy {

		t.Run(testName, func(t *testing.T) {

			env := &testEnv{args: test.input, headerArgs: test.headers}

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
		"should stop if -m is invalid":          "-m=FETCH http://foo",
		"should stop if -t is missing unit":     "-t=2 http://foo",
		"should stop if -H is invalid":          "-H=abc' http://foo",
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

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	gp "github.com/FabioSeixas/ogolpetai/ogolpetai"
)

const (
	bannerText = ` 
  ┌─┐  ┌─┐┌─┐┬  ┌─┐┌─┐  ┌┬┐┌─┐  ┌─┐┬
  │ │  │ ┬│ ││  ├─┘├┤    │ ├─┤  ├─┤│
  └─┘  └─┘└─┘┴─┘┴  └─┘   ┴ ┴ ┴  ┴ ┴┴
	  `

	usageText = `
Usage:
  ogolpetai [options] url
Options: `
)

func banner() string { return bannerText[1:] }

type flags struct {
	url       string
	n, c, rps int
	timeout   time.Duration
	method    string
	headers   []string
}

type ParseError int

const (
	NoArgs ParseError = iota
	WrongFormattedArg
	UnexpectedArg
	UnexpectedValue
	MissingRequiredArgument
	EmptyValue
)

type parseError struct {
	err   ParseError
	arg   *string
	value *string
}

func parseLegacy(f *flags) error {
	var _error parseError
	args := os.Args[1:]
	if len(args) < 1 {

		_error = parseError{err: NoArgs}
		goto error
	}

	for _, arg := range args {
		key, value, ok := strings.Cut(arg, "=")
		if !ok {
			_error = parseError{err: WrongFormattedArg, arg: &arg}
			goto error
		}

		if len(key) < 1 {
			_error = parseError{err: WrongFormattedArg, arg: &arg}
			goto error
		}

		if len(value) < 1 {
			_error = parseError{err: EmptyValue, arg: &arg}
			goto error
		}

		switch key {
		case "-c":
			v, err := strconv.Atoi(value)
			if err != nil {
				_error = parseError{err: UnexpectedValue, arg: &arg, value: &value}
				goto error
			}
			f.c = v
		case "-n":
			v, err := strconv.Atoi(value)
			if err != nil {
				_error = parseError{err: UnexpectedValue, arg: &arg, value: &value}
				goto error
			}
			f.n = v
		case "-url":
			v, err := url.Parse(value)
			if err != nil {
				_error = parseError{err: UnexpectedValue, arg: &arg, value: &value}
				goto error
			}
			f.url = v.String()
		default:
			_error = parseError{err: WrongFormattedArg, arg: &arg}
			goto error
		}

	}

	if len(f.url) < 1 {
		_error = parseError{err: MissingRequiredArgument}
		goto error
	}

	return nil

error:
	switch _error.err {
	case NoArgs:
		return errors.New("Missing arguments")
	case WrongFormattedArg:
		return errors.New(fmt.Sprintf("Wrong formatted argument: %q", *_error.arg))
	case UnexpectedArg:
		return errors.New(fmt.Sprintf("Unexpected argument: %q", *_error.arg))
	case UnexpectedValue:
		return errors.New(fmt.Sprintf("Unexpected value for argument %q: %q", *_error.arg, *_error.value))
	case EmptyValue:
		return errors.New(fmt.Sprintf("Empty value for argument: %q", *_error.arg))
	case MissingRequiredArgument:
		return errors.New(fmt.Sprintf("Required argument is missing"))
	default:
		return errors.New("UnknownError")
	}
}

type parseFunc func(string) error

func parseLegacyV2(f *flags) (err error) {

	parsers := map[string]parseFunc{
		"url": f.urlVar(&f.url),
		"n":   f.intVar(&f.n),
		"c":   f.intVar(&f.c),
	}

	for _, arg := range os.Args[1:] {
		key, value, ok := strings.Cut(arg, "=")
		if !ok {
			continue
		}
		parseFn, ok := parsers[strings.TrimPrefix(key, "-")]
		if !ok {
			continue
		}

		if err = parseFn(value); err != nil {
			err = fmt.Errorf("invalid value %q for flag %s: %w", value, key, err)
			break
		}
	}

	return err
}

func (f *flags) urlVar(p *string) parseFunc {
	return func(s string) error {
		_, err := url.Parse(s)
		*p = s
		return err
	}
}

func (f *flags) intVar(p *int) parseFunc {
	return func(s string) (err error) {
		*p, err = strconv.Atoi(s)
		return err
	}
}

type number int

func toNumber(p *int) *number {
	return (*number)(p)
}

func (n *number) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		err = errors.New("Parse error")
	} else if v <= 0 {
		err = errors.New("Should be positive")
	}

	*n = number(v)

	return err
}

func (n *number) String() string {
	return strconv.Itoa(int(*n))
}

type httpMethod string

func toHttpMethod(p *string) *httpMethod {
	return (*httpMethod)(p)
}

func (m *httpMethod) Set(s string) (err error) {

	switch s {
	case "GET", "POST", "PUT":
		*m = httpMethod(s)
	default:
		err = errors.New("Invalid http method")
	}

	return err
}

func (m *httpMethod) String() string {
	return string(*m)
}

type httpHeaders []string

func toHttpHeaders(p *[]string) *httpHeaders {
	return (*httpHeaders)(p)
}

func (h *httpHeaders) Set(s string) (err error) {

	_, _, ok := strings.Cut(s, ":")
	if !ok {
		err = errors.New("Wrong formatted header")
	}

	*h = append(*h, s)

	return err
}

func (h *httpHeaders) String() string {
	return strings.Join(*h, ", ")
}

func (f *flags) parse(s *flag.FlagSet, args []string) (err error) {
	s.Usage = func() {
		fmt.Fprintln(s.Output(), usageText)
		s.PrintDefaults()
	}

	// s.DurationVar(&f.timeout, "t", time.Duration(f.timeout), "Number of requests to make")
	s.Var(toNumber(&f.n), "n", "Number of requests to make")
	s.Var(toNumber(&f.c), "c", "Concurrency level")
	s.Var(toNumber(&f.rps), "t", "Throttle requests per second")
	s.Var(toHttpMethod(&f.method), "m", "Http method")
	s.Var(toHttpHeaders(&f.headers), "H", "Http header")

	if err = s.Parse(args); err != nil {
		return err
	}

	f.url = s.Arg(0)

	// fmt.Printf("%#v", f)

	if err = f.validate(); err != nil {
		fmt.Fprintln(s.Output(), err)
		s.Usage()
		return err
	}

	return nil
}

func (f *flags) validate() error {
	if err := validateURL(f.url); err != nil {
		return err
	}
	if f.n < f.c {
		return fmt.Errorf("-c=%d should be less or equal to -n=%d", f.c, f.n)
	}
	return nil
}

func validateURL(s string) error {
	u, err := url.Parse(s)

	switch {
	case strings.TrimSpace(s) == "":
		err = errors.New("'url' is Required")
	case err != nil:
		err = errors.New("parse error")
	case u.Scheme != "http":
		err = errors.New("only http protocol is supported")
	case u.Host == "":
		err = errors.New("Missing Host")
	}

	return err
}

func run(s *flag.FlagSet, args []string, out io.Writer) error {
	f := &flags{
		n:       100,
		c:       runtime.NumCPU(),
		rps:     0,
		timeout: time.Duration(10) * time.Second,
		method:  "GET",
		headers: []string{},
	}

	if err := f.parse(s, args); err != nil {
		return err
	}

	fmt.Fprintln(out, banner())
	fmt.Fprintf(
		out,
		"Making %d %s requests to %q with concurrency level %d (Timeout=%ds)\n",
		f.n,
		f.method,
		f.url,
		f.c,
		int(f.timeout.Seconds()),
	)

	if len(f.headers) > 0 {
		fmt.Fprintf(out, "Headers: %s", strings.Join(f.headers, ", "))

	}

	if f.rps > 0 {
		fmt.Fprintf(out, "(RPS: %d)\n", f.rps)
	}

	request, err := http.NewRequest(f.method, f.url, http.NoBody)

	if err != nil {
		return err
	}

	c := gp.Client{C: f.c, RPS: f.rps}
	sum := c.Do(request, f.n)
	sum.Fprint(out)

	return nil
}

func main() {
	if err := run(flag.CommandLine, os.Args[1:], os.Stdout); err != nil {
		os.Exit(1)
	}
}

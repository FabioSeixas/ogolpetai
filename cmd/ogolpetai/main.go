package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const (
	bannerText = ` 
  ┌─┐  ┌─┐┌─┐┬  ┌─┐┌─┐  ┌┬┐┌─┐  ┌─┐┬
  │ │  │ ┬│ ││  ├─┘├┤    │ ├─┤  ├─┤│
  └─┘  └─┘└─┘┴─┘┴  └─┘   ┴ ┴ ┴  ┴ ┴┴
	  `

	usageText = `
	  Usage:

	  -url
	  HTTP server URL to make requests (required)
	  -n
	  Number of requests to make
	  -c
	  Concurrency level`
)

func banner() string { return bannerText[1:] }
func usage() string  { return usageText[1:] }

type flags struct {
	url  string
	n, c int
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

func main() {
	f := &flags{
		n: 100,
		c: runtime.NumCPU(),
	}
	if err := parseLegacyV2(f); err != nil {
		fmt.Println(usage())
		log.Fatal(err)
		// os.Exit(0)
	}

	fmt.Println(banner())
	fmt.Printf("Making %d requests to %q with concurrency level %d\n", f.n, f.url, f.c)
}

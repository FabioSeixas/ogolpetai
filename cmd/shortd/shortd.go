package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	const (
		addr    = "localhost:8000"
		timeout = 10 * time.Second
	)

	fmt.Fprintf(os.Stdout, "starting the server on %s", addr)

	shortener := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "hello from the server")
		},
	)

	server := &http.Server{
		Addr:        addr,
		Handler:     http.TimeoutHandler(shortener, timeout, "request took too long"),
		ReadTimeout: timeout,
	}

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		fmt.Fprintln(os.Stderr, "server close", err)
	}
}

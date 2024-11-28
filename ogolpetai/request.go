package ogolpetai

import (
	"io"
	"net/http"
	"time"
)

type SendFunc func(*http.Request) *Result

func Send(r *http.Request) *Result {
	t := time.Now()
	var (
		code  int
		bytes int64
	)

	response, err := http.DefaultClient.Do(r)
	if err == nil {
		code = response.StatusCode
		bytes, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()

	}

	return &Result{
		Bytes:    bytes,
		Status:   code,
		Error:    err,
		Duration: time.Since(t),
	}
}

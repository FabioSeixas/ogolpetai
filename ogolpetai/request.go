package ogolpetai

import (
	"fmt"
	"net/http"
	"time"
)

type SendFunc func(*http.Request) *Result

func Send(r *http.Request) *Result {
	t := time.Now()
	fmt.Printf("request: %s\n", r.URL)
	time.Sleep(100 * time.Millisecond)

	return &Result{
		Bytes:    10,
		Status:   http.StatusOK,
		Duration: time.Since(t),
	}
}

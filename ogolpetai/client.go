package ogolpetai

import (
	"net/http"
	"time"
)

type Client struct {
}

func (c *Client) Do(r *http.Request, n int) *Result {
	t := time.Now()
	sum := c.do(r, n)
	return sum.Finalize(time.Since(t))
}

func (c *Client) do(r *http.Request, n int) *Result {
	var sum Result
	for ; n > 0; n-- {
		sum.Merge(Send(r))
	}
	return &sum
}

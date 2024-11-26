package ogolpetai

import (
	"context"
	"net/http"
	"time"
)

type Client struct {
	C   int // Concurrency level
	RPS int // RPS throttles the request per second
}

func (c *Client) Do(ctx context.Context, r *http.Request, n int) *Result {
	t := time.Now()
	sum := c.do(ctx, r, n)
	return sum.Finalize(time.Since(t))
}

func (c *Client) do(ctx context.Context, r *http.Request, n int) *Result {
	p := produce(ctx, n, func() req {
		return r.Clone(ctx)
	})

	if c.RPS > 0 {
		p = throttle(p, time.Second/time.Duration(c.RPS*c.C))
	}

	var sum Result
	for result := range split(p, c.C, Send) {
		sum.Merge(result)
	}
	return &sum
}

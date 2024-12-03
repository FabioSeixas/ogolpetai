package ogolpetai

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

type Option func(c *Client)

type Client struct {
	C       int           // Concurrency level
	RPS     int           // RPS throttles the request per second
	Timeout time.Duration // Timeout per request
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
		p = throttle(p, time.Second/time.Duration(c.RPS*c.concurrency()))
	}

	var (
		sum    Result
		client = c.client()
	)

	defer client.CloseIdleConnections()

	for result := range split(p, c.concurrency(), c.send(client)) {
		sum.Merge(result)
	}
	return &sum
}

func (c *Client) send(client *http.Client) SendFunc {
	return func(r *http.Request) *Result {
		return Send(client, r)
	}
}

func (c *Client) client() *http.Client {
	return &http.Client{
		Timeout: c.Timeout,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: c.concurrency(),
		},
	}

}

func (c *Client) concurrency() int {
	if c.C > 0 {
		return c.C
	}
	return runtime.NumCPU()
}

func Concurrency(c int) Option {
	return func(client *Client) {
		client.C = c
	}
}

func Timeout(t time.Duration) Option {
	return func(client *Client) {
		client.Timeout = t
	}
}

func RPS(rps int) Option {
	return func(client *Client) {
		client.RPS = rps
	}
}

func Do(ctx context.Context, url string, n int, opts ...Option) (*Result, error) {
	r, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("new http request: %w", err)
	}

	var c Client

	for _, opt := range opts {
		opt(&c)
	}

	return c.Do(ctx, r, n), nil
}

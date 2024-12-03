package ogolpetai

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type req = *http.Request

func Produce(ctx context.Context, out chan<- req, n int, fn func() req) {
	for ; n > 0; n-- {
		select {
		case <-ctx.Done():
			return
		case out <- fn():
		}

	}
}

func produce(ctx context.Context, n int, fn func() req) <-chan req {
	out := make(chan req)

	go func() {
		defer close(out)
		Produce(ctx, out, n, fn)
	}()

	return out
}

func Throttle(ctx context.Context, in <-chan req, out chan<- req, delay time.Duration) {
	t := time.NewTicker(delay)
	defer t.Stop()

	for {
		select {
		case r, ok := <-in:
			if !ok {
				return
			}
			<-t.C
			out <- r
		case <-ctx.Done():
			return
		}
	}
}

func throttle(ctx context.Context, in <-chan req, delay time.Duration) <-chan req {
	out := make(chan req)
	go func() {
		defer close(out)
		Throttle(ctx, in, out, delay)
	}()
	return out
}

func Split(in <-chan req, out chan<- *Result, c int, fn SendFunc) {
	send := func() {
		for r := range in {
			out <- fn(r)
		}
	}

	var wg sync.WaitGroup
	wg.Add(c)

	for ; c > 0; c-- {
		go func() {
			defer wg.Done()
			send()
		}()
	}
	wg.Wait()
}

func split(in <-chan req, c int, fn SendFunc) <-chan *Result {

	out := make(chan *Result)

	go func() {

		defer close(out)
		Split(in, out, c, fn)

	}()

	return out
}

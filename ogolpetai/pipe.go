package ogolpetai

import (
	"net/http"
	"sync"
	"time"
)

type req = *http.Request

func Produce(out chan<- req, n int, fn func() req) {
	for ; n > 0; n-- {
		out <- fn()
	}
}

func produce(n int, fn func() req) <-chan req {
	out := make(chan req)

	go func() {
		defer close(out)
		Produce(out, n, fn)
	}()

	return out
}

func Throttle(in <-chan req, out chan<- req, delay time.Duration) {
	t := time.NewTicker(delay)
	defer t.Stop()

	for r := range in {
		<-t.C
		out <- r
	}
}

func throttle(in <-chan req, delay time.Duration) <-chan req {
	out := make(chan req)
	go func() {
		defer close(out)
		Throttle(in, out, delay)
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

func split(in <-chan req, c int, fn SendFunc) chan<- *Result {

	out := make(chan *Result)

	go func() {

		defer close(out)
		Split(in, out, c, fn)

	}()

	return out
}

package ogolpetai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func newTestServer(t *testing.T, fn http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(fn))
	t.Cleanup(server.Close)
	return server
}

func newRequest(t *testing.T, server *httptest.Server) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, server.URL, http.NoBody)
	if err != nil {
		t.Fatalf("NewRequest err: %q; want nil", err)
	}
	return request
}

func TestClientDo(t *testing.T) {
	// t.Parallel()

	const wantHits, wantErrors = 10, 0
	var (
		gotHits atomic.Int64
		server  = newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			gotHits.Add(1)
		})
		request = newRequest(t, server)
	)

	c := &Client{
		C: 1,
	}

	sum := c.Do(context.Background(), request, wantHits)

	if got := gotHits.Load(); got != wantHits {
		t.Errorf("hits: %d; want: %d", got, wantHits)
	}

	if got := sum.Errors; got != wantErrors {
		t.Errorf("errors: %d; want: %d", got, wantHits)
	}
}

package http

import (
	"context"
	"net/http"
	"time"

	"github.com/beatlabs/patron/trace/http/cache"

	"github.com/beatlabs/patron/correlation"
	"github.com/beatlabs/patron/reliability/circuitbreaker"
	"github.com/beatlabs/patron/trace"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// Client interface of a HTTP client.
type Client interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

// TracedClient defines a HTTP client with tracing integrated.
type TracedClient struct {
	cl *http.Client
	cb *circuitbreaker.CircuitBreaker
	c  cache.Cache
}

// New creates a new HTTP client.
func New(oo ...OptionFunc) (*TracedClient, error) {
	tc := &TracedClient{
		cl: &http.Client{
			Timeout:   60 * time.Second,
			Transport: &nethttp.Transport{},
		},
		cb: nil,
	}

	for _, o := range oo {
		err := o(tc)
		if err != nil {
			return nil, err
		}
	}

	return tc, nil
}

// Do executes a HTTP request with integrated tracing and tracing propagation downstream.
// It also uses HTTP caching if it's enabled and skips tracing for these requests.
func (tc *TracedClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req,
		nethttp.OperationName(trace.HTTPOpName(req.Method, req.URL.String())),
		nethttp.ComponentName(trace.HTTPClientComponent))
	defer ht.Finish()

	req.Header.Set(correlation.HeaderID, correlation.IDFromContext(ctx))
	rsp, err := tc.send(req)
	if rsp.Header.Get(cache.XFromCache) != "1" {
		if err != nil {
			ext.Error.Set(ht.Span(), true)
		} else {
			ext.HTTPStatusCode.Set(ht.Span(), uint16(rsp.StatusCode))
		}

		ext.HTTPMethod.Set(ht.Span(), req.Method)
		ext.HTTPUrl.Set(ht.Span(), req.URL.String())
	}
	return rsp, err
}

func (tc *TracedClient) send(req *http.Request) (*http.Response, error) {
	if tc.c != nil {
		rsp, ok := tc.c.Get(*req)
		if !ok {
			rsp, err := tc.do(req)
			tc.c.Set(*req, rsp)
			return rsp, err
		}
		return rsp, nil
	}
	return tc.do(req)
}

func (tc *TracedClient) do(req *http.Request) (*http.Response, error) {
	if tc.cb == nil {
		return tc.cl.Do(req)
	}

	r, err := tc.cb.Execute(func() (interface{}, error) {
		return tc.cl.Do(req)
	})
	if err != nil {
		return nil, err
	}

	return r.(*http.Response), nil
}

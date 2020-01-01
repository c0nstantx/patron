package cache

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsFresh(t *testing.T) {
	now := time.Now().Unix()
	future := now + 10
	past := now - 10
	tests := []struct {
		name  string
		ce    cacheElement
		fresh bool
	}{
		{name: "fresh element", ce: cacheElement{expiresAt: future}, fresh: true},
		{name: "expired element", ce: cacheElement{expiresAt: past}, fresh: false},
		{name: "just expired element", ce: cacheElement{expiresAt: now}, fresh: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.fresh, isFresh(tt.ce))
		})
	}

}

func TestIsCacheable(t *testing.T) {
	getReq, _ := http.NewRequest("GET", "http://localhost", nil)
	headReq, _ := http.NewRequest("HEAD", "http://localhost", nil)
	postReq, _ := http.NewRequest("POST", "http://localhost", nil)
	successRsp := http.Response{StatusCode: 200}
	failedRsp := http.Response{StatusCode: 500}
	tests := []struct {
		name      string
		req       http.Request
		rsp       http.Response
		cacheable bool
	}{
		{name: "GET request with success response", req: *getReq, rsp: successRsp, cacheable: true},
		{name: "HEAD request with success response", req: *headReq, rsp: successRsp, cacheable: true},
		{name: "POST request with success response", req: *postReq, rsp: successRsp, cacheable: false},
		{name: "GET request with failed response", req: *getReq, rsp: failedRsp, cacheable: false},
		{name: "HEAD request with failed response", req: *headReq, rsp: failedRsp, cacheable: false},
		{name: "POST request with failed response", req: *postReq, rsp: failedRsp, cacheable: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.cacheable, isCacheable(tt.req, tt.rsp))
		})
	}

}

func TestRequestKey(t *testing.T) {
	getReq, _ := http.NewRequest("GET", "http://localhost", nil)
	headReq, _ := http.NewRequest("HEAD", "http://localhost", nil)
	tests := []struct {
		name     string
		req      http.Request
		expected string
	}{
		{name: "GET request", req: *getReq, expected: "GET http://localhost"},
		{name: "HEAD request", req: *headReq, expected: "HEAD http://localhost"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, requestKey(tt.req))
		})
	}
}

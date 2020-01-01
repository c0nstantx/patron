package cache

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	m, err := NewMemory(60)
	assert.Equal(t, int64(60), m.ttl)
	assert.Nil(t, err)
}

func TestGet(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost/existing", nil)
	misReq, _ := http.NewRequest("GET", "http://localhost/missing", nil)
	expReq, _ := http.NewRequest("GET", "http://localhost/expired", nil)
	errReq, _ := http.NewRequest("GET", "http://localhost/error", nil)
	rsp := &http.Response{StatusCode: 200}

	m, err := NewMemory(60)
	assert.Nil(t, err)
	m.Set(*req, rsp)
	m.store[requestKey(*expReq)] = cacheElement{expiresAt: time.Now().Unix() - 100, respBytes: []byte{}}
	m.store[requestKey(*errReq)] = cacheElement{expiresAt: time.Now().Unix() + 100, respBytes: []byte{155}}

	tests := []struct {
		name   string
		req    http.Request
		rsp    *http.Response
		exists bool
	}{
		{name: "existing in cache", req: *req, rsp: rsp, exists: true},
		{name: "missing in cache", req: *misReq, rsp: nil, exists: false},
		{name: "error in cache", req: *errReq, rsp: nil, exists: false},
		{name: "expired in cache", req: *expReq, rsp: nil, exists: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, ok := m.Get(tt.req)
			assert.Equal(t, tt.exists, ok)
			if ok {
				assert.Equal(t, tt.rsp.StatusCode, r.StatusCode)
			} else {
				assert.Nil(t, r)
			}
		})
	}
}

func TestSet(t *testing.T) {
	m, err := NewMemory(60)
	assert.Nil(t, err)
	getReq, _ := http.NewRequest("GET", "http://localhost", nil)
	getReq1, _ := http.NewRequest("GET", "http://localhost/1", nil)
	headReq, _ := http.NewRequest("HEAD", "http://localhost", nil)
	headReq1, _ := http.NewRequest("HEAD", "http://localhost/1", nil)
	postReq, _ := http.NewRequest("POST", "http://localhost", nil)
	postReq1, _ := http.NewRequest("POST", "http://localhost/1", nil)
	successRsp := http.Response{StatusCode: 200}
	failedRsp := http.Response{StatusCode: 500}
	tests := []struct {
		name   string
		req    http.Request
		rsp    http.Response
		cached bool
	}{
		{name: "GET request with success response", req: *getReq, rsp: successRsp, cached: true},
		{name: "HEAD request with success response", req: *headReq, rsp: successRsp, cached: true},
		{name: "POST request with success response", req: *postReq, rsp: successRsp, cached: false},
		{name: "GET request with failed response", req: *getReq1, rsp: failedRsp, cached: false},
		{name: "HEAD request with failed response", req: *headReq1, rsp: failedRsp, cached: false},
		{name: "POST request with failed response", req: *postReq1, rsp: failedRsp, cached: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.Set(tt.req, &tt.rsp)
			_, ok := m.Get(tt.req)
			assert.Equal(t, tt.cached, ok)
		})
	}

}

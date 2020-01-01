package cache

import (
	"bufio"
	"bytes"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/beatlabs/patron/log"
	"github.com/pkg/errors"
)

// Memory is the cache layer struct for in-memory caching
type Memory struct {
	sync.Mutex
	store map[string]cacheElement
	ttl   int64
}

// NewMemory creates a new in-memory cache
func NewMemory(ttl int64) (*Memory, error) {
	return &Memory{
		ttl:   ttl,
		store: make(map[string]cacheElement),
	}, nil
}

// Get retrieves a value from cache if exists.
func (m *Memory) Get(req http.Request) (*http.Response, bool) {
	m.Lock()
	ce, ok := m.store[requestKey(req)]
	m.Unlock()
	if !ok {
		return nil, false
	}
	if !isFresh(ce) {
		m.Delete(req)
		return nil, false
	}
	b := bytes.NewBuffer(ce.respBytes)
	rsp, err := http.ReadResponse(bufio.NewReader(b), &req)
	if err != nil {
		err := errors.Wrapf(err, "Error reading cached response for request: %s %s", req.Method, req.URL.String())
		log.Warn(err)
		m.Delete(req)
		return nil, false
	}
	rsp.Header.Set(XFromCache, "1")
	return rsp, true
}

// Set creates/updates a value in cache.
func (m *Memory) Set(req http.Request, rsp *http.Response) {
	if isCacheable(req, *rsp) {
		respBytes, err := httputil.DumpResponse(rsp, true)
		if err == nil {
			respText := string(respBytes)
			ce := cacheElement{
				respBytes: []byte(respText),
				expiresAt: time.Now().Unix() + m.ttl,
			}
			m.Lock()
			m.store[requestKey(req)] = ce
			m.Unlock()
		}
	}
}

// Delete removes a value from cache.
func (m *Memory) Delete(req http.Request) {
	m.Lock()
	delete(m.store, requestKey(req))
	m.Unlock()
}

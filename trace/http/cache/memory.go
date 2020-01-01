package cache

import (
	"bufio"
	"bytes"
	"io/ioutil"
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
	defer m.Unlock()
	if ce, ok := m.store[requestKey(req)]; ok {
		if !isFresh(ce) {
			m.Delete(req)
			return nil, false
		}
		b := bytes.NewBuffer(ce.respBytes)
		rsp, err := http.ReadResponse(bufio.NewReader(b), &req)
		if err != nil {
			err := errors.Wrapf(err, "Error reading cached response for request: %s %s", req.Method, req.URL.String())
			log.Warn(err)
			m.Unlock()
			m.Delete(req)
			m.Lock()
			return nil, false
		}
		return rsp, true
	}
	return nil, false
}

// Set creates/updates a value in cache.
func (m *Memory) Set(req http.Request, rsp *http.Response) {
	m.Lock()
	if isCacheable(req, *rsp) {
		// respBytes, err := ioutil.ReadAll(rsp.Body)
		respBytes, err := httputil.DumpResponse(rsp, true)
		if err == nil {
			respText := string(respBytes)
			ce := cacheElement{
				respBytes: []byte(respText),
				expiresAt: time.Now().Unix() + m.ttl,
			}
			m.store[requestKey(req)] = ce
			rsp.Body = ioutil.NopCloser(bytes.NewReader(ce.respBytes))
		}
	}
	m.Unlock()
}

// Delete removes a value from cache.
func (m *Memory) Delete(req http.Request) {
	m.Lock()
	delete(m.store, requestKey(req))
	m.Unlock()
}

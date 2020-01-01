package cache

import (
	"fmt"
	"net/http"
	"time"
)

// Cache is the definition of cache layer based on HTTP caching (https://tools.ietf.org/html/rfc7234)
type Cache interface {
	Get(req http.Request) (rsp *http.Response, ok bool)
	Set(req http.Request, rsp *http.Response)
	Delete(req http.Request)
}

type cacheElement struct {
	expiresAt int64
	respBytes []byte
}

func requestKey(req http.Request) string {
	return fmt.Sprintf("%s %s", req.Method, req.URL.String())
}

func isCacheable(req http.Request, rsp http.Response) bool {
	// Cache only GET and HEAD requests that returned OK.
	// Incomplete responses (206: Partial Content) won't be cached for the moment.
	reqCacheable := req.Method == "GET" || req.Method == "HEAD"
	rspCacheable := rsp.StatusCode == 200
	return reqCacheable && rspCacheable
}

func isFresh(ce cacheElement) bool {
	return ce.expiresAt > time.Now().Unix()
}

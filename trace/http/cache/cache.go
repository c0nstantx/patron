package cache

import (
	"fmt"
	"net/http"
	"time"
)

// XFromCache is the header added to responses that are returned from the cache
const XFromCache = "X-From-Cache"

// Cache is the definition of cache layer based on HTTP caching (https://tools.ietf.org/html/rfc7234)
// Currently only successful GET and HEAD requests (200: OK) are cached.
// TODO: Handle incomplete responses (206: Partial Content)
// TODO: Handle authenticated requests
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
	reqCacheable := req.Method == "GET" || req.Method == "HEAD"
	rspCacheable := rsp.StatusCode == 200
	return reqCacheable && rspCacheable
}

func isFresh(ce cacheElement) bool {
	return ce.expiresAt > time.Now().Unix()
}

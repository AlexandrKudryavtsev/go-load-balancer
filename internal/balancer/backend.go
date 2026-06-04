package balancer

import (
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Backend struct {
	URL        *url.URL
	Alive      atomic.Bool
	Proxy      *httputil.ReverseProxy
	HealthPath string
}

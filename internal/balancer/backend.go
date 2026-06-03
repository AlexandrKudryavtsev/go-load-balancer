package balancer

import (
	"net/http/httputil"
	"net/url"
)

type Backend struct {
	URL   *url.URL
	Alive bool
	Proxy *httputil.ReverseProxy
}

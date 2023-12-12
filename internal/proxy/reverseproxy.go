package proxy

import (
	. "github.com/CharLemAznable/violet/internal/types"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewReverseProxy(target *url.URL) ReverseProxy {
	reverseProxy := httputil.NewSingleHostReverseProxy(target)
	director := reverseProxy.Director
	reverseProxy.Director = func(req Req) {
		req.Host = target.Host // set header Host
		director(req)
	}
	reverseProxy.Transport = RoundTripperFunc(defaultRoundTrip)
	return reverseProxy
}

var pRoundTrip = http.DefaultTransport.RoundTrip

func defaultRoundTrip(req Req) (Rsp, error) {
	return pRoundTrip(req)
}

func EnableMockRoundTrip(mock func(Req) (Rsp, error)) {
	pRoundTrip = mock
}

func DisableMockRoundTrip() {
	pRoundTrip = http.DefaultTransport.RoundTrip
}

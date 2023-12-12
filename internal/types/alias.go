package types

import (
	"net/http"
	"net/http/httputil"
)

type Req = *http.Request

type Rsp = *http.Response

type ReverseProxy = *httputil.ReverseProxy

type ReverseProxyDecorator = func(ReverseProxy) ReverseProxy

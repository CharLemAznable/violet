package resilience_test

import (
	"bufio"
	"errors"
	"github.com/CharLemAznable/gogo/ext"
	"github.com/CharLemAznable/gogo/lang"
	. "github.com/CharLemAznable/violet/internal/proxy"
	"github.com/CharLemAznable/violet/internal/resilience"
	. "github.com/CharLemAznable/violet/internal/types"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestFallbackNotEnabled(t *testing.T) {
	decorator := resilience.NewFallbackPlugin(&resilience.FallbackConfig{})
	if !lang.Equal(ReverseProxyIdentity, decorator.Decorator) {
		t.Error("Expected get ReverseProxyIdentity but not")
	}

	decorator = resilience.NewFallbackPlugin(&resilience.FallbackConfig{Enabled: "true"})
	if !lang.Equal(ReverseProxyIdentity, decorator.Decorator) {
		t.Error("Expected get ReverseProxyIdentity but not")
	}
}

func TestFallbackResponse(t *testing.T) {
	decorator := resilience.NewDecorator("test", newFallbackResilienceConfig(&resilience.FallbackConfig{
		Enabled:          "true",
		FallbackResponse: "HTTP/1.1 200 OK\r\n\r\nFallbackResponse",
	}))
	EnableMockRoundTrip(func(req Req) (Rsp, error) {
		return nil, errors.New("error")
	})
	backendURL, _ := url.Parse("http://a.b.c")
	reverseProxy := decorator.Decorate(NewReverseProxy(backendURL))
	frontend := httptest.NewServer(reverseProxy)

	request, _ := http.NewRequest("GET", frontend.URL, nil)
	response, _ := frontend.Client().Do(request)
	responseBody, _ := ext.DumpResponseBody(response)
	if string(responseBody) != "FallbackResponse" {
		t.Errorf("Expected responseBody is 'FallbackResponse', but got '%s'", string(responseBody))
	}
	DisableMockRoundTrip()
}

func TestFallbackFunction(t *testing.T) {
	_ = resilience.RegisterFallbackFunction("demo_func",
		func(req Req, context map[string]string) (Rsp, error) {
			return http.ReadResponse(bufio.NewReader(
				strings.NewReader("HTTP/1.1 200 OK\r\n\r\nFallbackFunction")), req)
		})

	decorator := resilience.NewDecorator("test", newFallbackResilienceConfig(&resilience.FallbackConfig{
		Enabled:          "true",
		FallbackFunction: "demo_func",
	}))
	EnableMockRoundTrip(func(req Req) (Rsp, error) {
		panic("panic")
	})
	backendURL, _ := url.Parse("http://a.b.c")
	reverseProxy := decorator.Decorate(NewReverseProxy(backendURL))
	frontend := httptest.NewServer(reverseProxy)

	request, _ := http.NewRequest("GET", frontend.URL, nil)
	response, _ := frontend.Client().Do(request)
	responseBody, _ := ext.DumpResponseBody(response)
	if string(responseBody) != "FallbackFunction" {
		t.Errorf("Expected responseBody is 'FallbackFunction', but got '%s'", string(responseBody))
	}
	DisableMockRoundTrip()
}

func TestFallback(t *testing.T) {
	EnableMockRoundTrip(func(req Req) (Rsp, error) {
		return http.ReadResponse(bufio.NewReader(
			strings.NewReader("HTTP/1.1 500 Internal Server Error\r\n\r\n")), req)
	})
	decorator := resilience.NewDecorator("test", newFallbackResilienceConfig(&resilience.FallbackConfig{
		Enabled:          "true",
		FallbackResponse: "HTTP/1.1 200 OK\r\n\r\nFallbackResponse",
		FallbackFunction: "not_exists_func",
	}))
	backendURL, _ := url.Parse("http://a.b.c")
	reverseProxy := decorator.Decorate(NewReverseProxy(backendURL))
	frontend := httptest.NewServer(reverseProxy)

	request, _ := http.NewRequest("GET", frontend.URL, nil)
	response, _ := frontend.Client().Do(request)
	responseBody, _ := ext.DumpResponseBody(response)
	if string(responseBody) != "FallbackResponse" {
		t.Errorf("Expected responseBody is 'FallbackResponse', but got '%s'", string(responseBody))
	}
	DisableMockRoundTrip()
}

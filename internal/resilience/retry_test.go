package resilience_test

import (
	"bufio"
	"github.com/CharLemAznable/gogo/ext"
	. "github.com/CharLemAznable/violet/internal/proxy"
	"github.com/CharLemAznable/violet/internal/resilience"
	. "github.com/CharLemAznable/violet/internal/types"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRetry(t *testing.T) {
	entry, _ := resilience.NewRetryPlugin("test", &resilience.RetryConfig{Disabled: "true"})
	if entry != nil {
		t.Errorf("Expected retry is nil, but got '%v'", entry)
	}

	decorator := resilience.NewDecorator("test", newRetryResilienceConfig(&resilience.RetryConfig{
		MaxAttempts:            "2",
		FailAfterMaxAttempts:   "true",
		WaitInterval:           "0",
		WhenMaxRetriesResponse: "HTTP/1.1 200 OK\r\n\r\nRetryMaxRetries",
	}))
	EnableMockRoundTrip(func(req Req) (Rsp, error) {
		return http.ReadResponse(bufio.NewReader(
			strings.NewReader("HTTP/1.1 500 Internal Server Error\r\n\r\n")), req)
	})
	backendURL, _ := url.Parse("http://a.b.c")
	reverseProxy := decorator.Decorate(NewReverseProxy(backendURL))
	frontend := httptest.NewServer(reverseProxy)

	request, _ := http.NewRequest("GET", frontend.URL, nil)
	response, _ := frontend.Client().Do(request)
	responseBody, _ := ext.DumpResponseBody(response)
	if string(responseBody) != "RetryMaxRetries" {
		t.Errorf("Expected responseBody is 'RetryMaxRetries', but got '%s'", string(responseBody))
	}

	DisableMockRoundTrip()
}

package resilience_test

import (
	. "github.com/CharLemAznable/violet/internal/elf"
	. "github.com/CharLemAznable/violet/internal/proxy"
	"github.com/CharLemAznable/violet/internal/resilience"
	. "github.com/CharLemAznable/violet/internal/types"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestTimeLimiter(t *testing.T) {
	entry, _ := resilience.NewTimeLimiterPlugin("test", &resilience.TimeLimiterConfig{Disabled: "true"})
	if entry != nil {
		t.Errorf("Expected timelimiter is nil, but got '%v'", entry)
	}
	entry, _ = resilience.NewTimeLimiterPlugin("test", &resilience.TimeLimiterConfig{})
	if entry == nil {
		t.Errorf("Expected timelimiter is not nil, but got nil")
	}

	decorator := resilience.NewDecorator("test", newTimeLimiterResilienceConfig(&resilience.TimeLimiterConfig{
		TimeoutDuration:     "1s",
		WhenTimeoutResponse: "HTTP/1.1 200 OK\r\n\r\nTimeLimiterTimeout",
	}))
	backend := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r Req) {
			time.Sleep(time.Second * 2)
			_, _ = w.Write([]byte("success"))
		}))
	backendURL, _ := url.Parse(backend.URL)
	reverseProxy := decorator.Decorate(NewReverseProxy(backendURL))
	frontend := httptest.NewServer(reverseProxy)

	request, _ := http.NewRequest("GET", frontend.URL, nil)
	response, _ := frontend.Client().Do(request)
	responseBody, _ := DumpResponseBody(response)
	if string(responseBody) != "TimeLimiterTimeout" {
		t.Errorf("Expected responseBody is 'TimeLimiterTimeout', but got '%s'", string(responseBody))
	}
}

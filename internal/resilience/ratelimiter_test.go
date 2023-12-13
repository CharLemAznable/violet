package resilience_test

import (
	. "github.com/CharLemAznable/violet/internal/elf"
	. "github.com/CharLemAznable/violet/internal/proxy"
	"github.com/CharLemAznable/violet/internal/resilience"
	. "github.com/CharLemAznable/violet/internal/types"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	entry, _ := resilience.NewRateLimiterPlugin("test", &resilience.RateLimiterConfig{Disabled: "true"})
	if entry != nil {
		t.Errorf("Expected ratelimiter is nil, but got '%v'", entry)
	}

	decorator := resilience.NewDecorator("test", newRateLimiterResilienceConfig(&resilience.RateLimiterConfig{
		TimeoutDuration:      "1s",
		LimitRefreshPeriod:   "2s",
		LimitForPeriod:       "2",
		WhenOverRateResponse: "HTTP/1.1 200 OK\r\n\r\nRateLimiterNotPermitted",
	}))
	backend := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r Req) {
			time.Sleep(time.Millisecond * 500)
			_, _ = w.Write([]byte("success"))
		}))
	backendURL, _ := url.Parse(backend.URL)
	reverseProxy := decorator.Decorate(NewReverseProxy(backendURL))
	frontend := httptest.NewServer(reverseProxy)

	request, _ := http.NewRequest("GET", frontend.URL, nil)
	responses := []chan Rsp{
		make(chan Rsp, 1),
		make(chan Rsp, 1),
		make(chan Rsp, 1),
	}
	for i := 0; i < 3; i++ {
		go func(i int) {
			response, _ := frontend.Client().Do(request)
			responses[i] <- response
		}(i)
	}
	response0 := <-responses[0]
	response1 := <-responses[1]
	response2 := <-responses[2]
	responseBody0, _ := DumpResponseBody(response0)
	responseBody1, _ := DumpResponseBody(response1)
	responseBody2, _ := DumpResponseBody(response2)
	bodies := []string{string(responseBody0), string(responseBody1), string(responseBody2)}
	sort.Strings(bodies)
	if bodies[0] != "RateLimiterNotPermitted" {
		t.Errorf("Expected responseBody[0] has 'RateLimiterNotPermitted', but got '%s'", bodies[0])
	}
	if bodies[1] != "success" {
		t.Errorf("Expected responseBody[1] has 'success', but got '%s'", bodies[1])
	}
	if bodies[2] != "success" {
		t.Errorf("Expected responseBody[2] has 'success', but got '%s'", bodies[2])
	}
}

package resilience_test

import (
	"bufio"
	"errors"
	"github.com/CharLemAznable/gogo/ext"
	. "github.com/CharLemAznable/violet/internal/proxy"
	"github.com/CharLemAznable/violet/internal/resilience"
	. "github.com/CharLemAznable/violet/internal/types"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	entry, _ := resilience.NewCircuitBreakerPlugin("test", &resilience.CircuitBreakerConfig{Disabled: "true"})
	if entry != nil {
		t.Errorf("Expected circuitbreaker is nil, but got '%v'", entry)
	}

	decorator := resilience.NewDecorator("test", newCircuitBreakerResilienceConfig(&resilience.CircuitBreakerConfig{
		SlidingWindowSize:                     "10",
		MinimumNumberOfCalls:                  "10",
		FailureRateThreshold:                  "50",
		AutomaticTransitionFromOpenToHalfOpen: "true",
		WaitIntervalInOpenState:               "5s",
		PermittedNumberOfCallsInHalfOpenState: "2",
		WhenOverLoadResponse:                  "HTTP/1.1 200 OK\r\n\r\nCircuitBreakerNotPermitted",
	}))
	EnableMockRoundTrip(func(req Req) (Rsp, error) {
		body, _ := ext.DumpRequestBody(req)
		i, _ := strconv.ParseInt(string(body), 10, 64)
		if i%2 == 0 {
			return http.ReadResponse(bufio.NewReader(
				strings.NewReader("HTTP/1.1 200 OK\r\n\r\nOK")), req)
		}
		return nil, errors.New("error")
	})
	backendURL, _ := url.Parse("http://a.b.c")
	reverseProxy := decorator.Decorate(NewReverseProxy(backendURL))
	reverseProxy.ErrorHandler = func(rw http.ResponseWriter, _ Req, _ error) {
		rw.WriteHeader(http.StatusBadGateway)
	}
	frontend := httptest.NewServer(reverseProxy)

	var wg sync.WaitGroup
	// 启动多个协程
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			request, _ := http.NewRequest("GET", frontend.URL, strings.NewReader(strconv.Itoa(i)))
			_, _ = frontend.Client().Do(request)
		}(i)
	}
	// 等待所有协程执行完毕
	wg.Wait()

	request, _ := http.NewRequest("GET", frontend.URL, strings.NewReader(strconv.Itoa(10)))
	response, _ := frontend.Client().Do(request)
	responseBody, _ := ext.DumpResponseBody(response)
	if string(responseBody) != "CircuitBreakerNotPermitted" {
		t.Errorf("Expected responseBody is 'CircuitBreakerNotPermitted', but got '%s'", string(responseBody))
	}

	DisableMockRoundTrip()
}

func TestCircuitBreakerSlow(t *testing.T) {
	decorator := resilience.NewDecorator("test", newCircuitBreakerResilienceConfig(&resilience.CircuitBreakerConfig{
		SlidingWindowType:                     "time_based",
		SlidingWindowSize:                     "10",
		MinimumNumberOfCalls:                  "10",
		SlowCallRateThreshold:                 "100",
		SlowCallDurationThreshold:             "1s",
		PermittedNumberOfCallsInHalfOpenState: "2",
		MaxWaitDurationInHalfOpenState:        "5s",
	}))
	EnableMockRoundTrip(func(req Req) (Rsp, error) {
		time.Sleep(time.Second * 2)
		return http.ReadResponse(bufio.NewReader(
			strings.NewReader("HTTP/1.1 200 OK\r\n\r\nOK")), req)
	})
	backendURL, _ := url.Parse("http://a.b.c")
	reverseProxy := decorator.Decorate(NewReverseProxy(backendURL))
	reverseProxy.ErrorHandler = func(rw http.ResponseWriter, _ Req, _ error) {
		rw.WriteHeader(http.StatusBadGateway)
	}
	frontend := httptest.NewServer(reverseProxy)

	var wg sync.WaitGroup
	// 启动多个协程
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			request, _ := http.NewRequest("GET", frontend.URL, nil)
			_, _ = frontend.Client().Do(request)
		}(i)
	}
	// 等待所有协程执行完毕
	wg.Wait()

	request, _ := http.NewRequest("GET", frontend.URL, nil)
	response, _ := frontend.Client().Do(request)
	if http.StatusBadGateway != response.StatusCode {
		t.Errorf("Expected StatusCode is '502', but got '%d'", response.StatusCode)
	}

	DisableMockRoundTrip()
}

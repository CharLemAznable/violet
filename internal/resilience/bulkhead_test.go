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

func TestBulkhead(t *testing.T) {
	entry, _ := resilience.NewBulkheadPlugin("test", &resilience.BulkheadConfig{Disabled: "true"})
	if entry != nil {
		t.Errorf("Expected bulkhead is nil, but got '%v'", entry)
	}

	_, decorator := resilience.NewBulkheadPlugin("test", &resilience.BulkheadConfig{
		MaxConcurrentCalls: "1",
		MaxWaitDuration:    "1s",
		WhenFullResponse:   "HTTP/1.1 200 OK\r\n\r\nBulkheadFull",
	})
	backend := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r Req) {
			time.Sleep(time.Second * 3)
			_, _ = w.Write([]byte("success"))
		}))
	backendURL, _ := url.Parse(backend.URL)
	reverseProxy := decorator(NewReverseProxy(backendURL))
	frontend := httptest.NewServer(reverseProxy)

	request, _ := http.NewRequest("GET", frontend.URL, nil)
	resp1 := make(chan Rsp, 1)
	resp2 := make(chan Rsp, 1)
	go func() {
		response, _ := frontend.Client().Do(request)
		resp1 <- response
	}()
	time.Sleep(time.Second * 1)
	go func() {
		response, _ := frontend.Client().Do(request)
		resp2 <- response
	}()
	response1 := <-resp1
	response2 := <-resp2
	responseBody1, _ := DumpResponseBody(response1)
	if string(responseBody1) != "success" {
		t.Errorf("Expected responseBody1 is 'success', but got '%s'", string(responseBody1))
	}
	responseBody2, _ := DumpResponseBody(response2)
	if string(responseBody2) != "BulkheadFull" {
		t.Errorf("Expected responseBody2 is 'BulkheadFull', but got '%s'", string(responseBody2))
	}
}

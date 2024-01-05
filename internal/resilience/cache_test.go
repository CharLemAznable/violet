package resilience_test

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/CharLemAznable/gogo/ext"
	. "github.com/CharLemAznable/violet/internal/proxy"
	"github.com/CharLemAznable/violet/internal/resilience"
	. "github.com/CharLemAznable/violet/internal/types"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	entry, _ := resilience.NewCachePlugin("test", &resilience.CacheConfig{})
	if entry != nil {
		t.Errorf("Expected cache is nil, but got '%v'", entry)
	}

	decorator := resilience.NewDecorator("test", newCacheResilienceConfig(&resilience.CacheConfig{
		Enabled:  "true",
		Capacity: "10",
		ItemTTL:  "10s",
	}))
	EnableMockRoundTrip(func(req Req) (Rsp, error) {
		code := req.URL.Query().Get("code")
		statusCode, _ := strconv.Atoi(code)
		statusText := http.StatusText(statusCode)
		if statusText == "" {
			return nil, errors.New("code error")
		}
		resp := fmt.Sprintf("HTTP/1.1 %s %s\r\n\r\n%s", code, statusText, randString(4))
		return http.ReadResponse(bufio.NewReader(strings.NewReader(resp)), req)
	})
	backendURL, _ := url.Parse("http://a.b.c")
	reverseProxy := decorator.Decorate(NewReverseProxy(backendURL))
	frontend := httptest.NewServer(reverseProxy)

	request, _ := http.NewRequest("GET", frontend.URL+"?code=200", nil)
	response, _ := frontend.Client().Do(request)
	responseBody1, _ := ext.DumpResponseBody(response)
	request, _ = http.NewRequest("GET", frontend.URL+"?code=200", nil)
	response, _ = frontend.Client().Do(request)
	responseBody2, _ := ext.DumpResponseBody(response)
	if string(responseBody1) != string(responseBody2) {
		t.Errorf("Expected code 200 cached, but got '%s' and '%s'", string(responseBody1), string(responseBody2))
	}

	request, _ = http.NewRequest("GET", frontend.URL+"?code=301", nil)
	response, _ = frontend.Client().Do(request)
	responseBody1, _ = ext.DumpResponseBody(response)
	request, _ = http.NewRequest("GET", frontend.URL+"?code=301", nil)
	response, _ = frontend.Client().Do(request)
	responseBody2, _ = ext.DumpResponseBody(response)
	if string(responseBody1) != string(responseBody2) {
		t.Errorf("Expected code 301 cached, but got '%s' and '%s'", string(responseBody1), string(responseBody2))
	}

	request, _ = http.NewRequest("GET", frontend.URL+"?code=404", nil)
	response, _ = frontend.Client().Do(request)
	responseBody1, _ = ext.DumpResponseBody(response)
	request, _ = http.NewRequest("GET", frontend.URL+"?code=404", nil)
	response, _ = frontend.Client().Do(request)
	responseBody2, _ = ext.DumpResponseBody(response)
	if string(responseBody1) != string(responseBody2) {
		t.Errorf("Expected code 404 cached, but got '%s' and '%s'", string(responseBody1), string(responseBody2))
	}

	request, _ = http.NewRequest("GET", frontend.URL+"?code=500", nil)
	response, _ = frontend.Client().Do(request)
	responseBody1, _ = ext.DumpResponseBody(response)
	request, _ = http.NewRequest("GET", frontend.URL+"?code=500", nil)
	response, _ = frontend.Client().Do(request)
	responseBody2, _ = ext.DumpResponseBody(response)
	if string(responseBody1) == string(responseBody2) {
		t.Errorf("Expected code 500 not cached, but got '%s' and '%s'", string(responseBody1), string(responseBody2))
	}

	request, _ = http.NewRequest("GET", frontend.URL+"?code=1", nil)
	response, _ = frontend.Client().Do(request)
	if http.StatusBadGateway != response.StatusCode {
		t.Errorf("Expected StatusCode is '502', but got '%d'", response.StatusCode)
	}

	DisableMockRoundTrip()
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[random.Intn(len(letterRunes))]
	}
	return string(b)
}

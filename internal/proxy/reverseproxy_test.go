package proxy_test

import (
	"bufio"
	"github.com/CharLemAznable/gogo/ext"
	"github.com/CharLemAznable/violet/internal/proxy"
	. "github.com/CharLemAznable/violet/internal/types"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewReverseProxy(t *testing.T) {
	subscriber := ext.SubFn(logDumpMessage)
	ext.Subscribe(proxy.DumpTopic, subscriber)
	body := "dump content"
	testServer := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r Req) {
			w.Header().Set("ActualHost", r.Host)
			_, _ = w.Write([]byte(body))
		}))
	testServerURL, _ := url.Parse(testServer.URL)

	reverseProxy := proxy.NewReverseProxy(testServerURL)
	proxy.DumpDecorator(true, proxy.TargetDump, "test")(
		proxy.DumpDecorator(false, proxy.SourceDump, "test")(reverseProxy))
	reverseServer := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r Req) {
			reverseProxy.ServeHTTP(w, r)
		}))

	request, _ := http.NewRequest("GET", reverseServer.URL, strings.NewReader(body))
	response, _ := reverseServer.Client().Do(request)
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status code is 200, but got '%d'", response.StatusCode)
	}
	actualHost := response.Header.Get("ActualHost")
	if actualHost != testServerURL.Host {
		t.Errorf("Expected actual host is '%s', but got '%s'", testServerURL.Host, actualHost)
	}
	responseBody, _ := ext.DumpResponseBody(response)
	if body != string(responseBody) {
		t.Errorf("Expected responseBody is '%s', but got '%s'", body, string(responseBody))
	}

	proxy.EnableMockRoundTrip(func(req Req) (Rsp, error) {
		return http.ReadResponse(bufio.NewReader(
			strings.NewReader("HTTP/1.1 200 OK\r\n\r\nOK")), req)
	})
	request, _ = http.NewRequest("GET", reverseServer.URL, nil)
	response, _ = reverseServer.Client().Do(request)
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status code is 200, but got '%d'", response.StatusCode)
	}
	responseBody, _ = ext.DumpResponseBody(response)
	if "OK" != string(responseBody) {
		t.Errorf("Expected responseBody is 'OK', but got '%s'", string(responseBody))
	}
	proxy.DisableMockRoundTrip()
	time.Sleep(time.Second)
	ext.Unsubscribe(proxy.DumpTopic, subscriber)
}

var dumpLogger = log.New(os.Stdout, "violet >> ", log.LstdFlags|log.Lmsgprefix)

func logDumpMessage(message *proxy.DumpMessage) {
	if message.Error != nil {
		dumpLogger.Printf("%d: %s %s error: %v\n", message.UnixNano,
			message.Name, message.MessageType, message.Error)
	} else {
		dumpLogger.Printf("%d: %s %s:\n%s\n", message.UnixNano,
			message.Name, message.MessageType, string(message.Content))
	}
}

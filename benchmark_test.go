package violet_test

import (
	"github.com/CharLemAznable/violet"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkServeMux(b *testing.B) {
	mux := newServeMux("OK")
	server := httptest.NewServer(mux)
	for i := 0; i < b.N; i++ {
		resp, _ := server.Client().Get(server.URL)
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "OK" {
			b.Errorf("Expected ret is 'OK', but got '%v'", string(body))
		}
	}
}

type ServeMuxProxy struct {
	mux atomic.Pointer[http.ServeMux]
}

func NewServeMuxProxy(mux *http.ServeMux) *ServeMuxProxy {
	proxy := &ServeMuxProxy{}
	proxy.mux.Store(mux)
	return proxy
}

func (proxy *ServeMuxProxy) Update(mux *http.ServeMux) {
	proxy.mux.Store(mux)
}

func (proxy *ServeMuxProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxy.loadMux().ServeHTTP(w, r)
}

func (proxy *ServeMuxProxy) loadMux() *http.ServeMux {
	return proxy.mux.Load()
}

func BenchmarkServeMuxProxy_Simple(b *testing.B) {
	mux := newServeMux("OK")
	serveProxy := NewServeMuxProxy(mux)
	server := httptest.NewServer(serveProxy)
	for i := 0; i < b.N; i++ {
		resp, _ := server.Client().Get(server.URL)
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "OK" {
			b.Errorf("Expected ret is 'OK', but got '%v'", string(body))
		}
	}
}

func BenchmarkServeMuxProxy_Routine(b *testing.B) {
	mux1 := newServeMux("OK")
	mux2 := newServeMux("OK")
	serveProxy := NewServeMuxProxy(mux1)
	var done atomic.Bool
	go func() {
		for !done.Load() {
			serveProxy.Update(mux2)
			mux1, mux2 = mux2, mux1
			time.Sleep(time.Millisecond)
		}
	}()
	server := httptest.NewServer(serveProxy)
	for i := 0; i < b.N; i++ {
		resp, _ := server.Client().Get(server.URL)
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "OK" {
			b.Errorf("Expected ret is 'OK', but got '%v'", string(body))
		}
	}
	done.Store(true)
}

func BenchmarkDataPlane_Simple(b *testing.B) {
	mux := newServeMux("OK")
	backend := httptest.NewServer(mux)
	dataPlane := violet.NewDataPlane(&violet.Config{
		Endpoint: []violet.EndpointConfig{
			{Name: "benchmark", Location: "/", TargetURL: backend.URL},
		},
	})
	frontend := httptest.NewServer(dataPlane)
	for i := 0; i < b.N; i++ {
		resp, _ := frontend.Client().Get(frontend.URL)
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "OK" {
			b.Errorf("Expected ret is 'OK', but got '%v'", string(body))
		}
	}
}

func BenchmarkDataPlane_Cache(b *testing.B) {
	mux := newServeMux("OK")
	backend := httptest.NewServer(mux)
	dataPlane := violet.NewDataPlane(&violet.Config{
		Endpoint: []violet.EndpointConfig{
			{
				Name:      "benchmark",
				Location:  "/",
				TargetURL: backend.URL,
				Resilience: violet.ResilienceConfig{
					Cache: violet.CacheConfig{Enabled: "true"},
				},
			},
		},
	})
	frontend := httptest.NewServer(dataPlane)
	for i := 0; i < b.N; i++ {
		resp, _ := frontend.Client().Get(frontend.URL)
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "OK" {
			b.Errorf("Expected ret is 'OK', but got '%v'", string(body))
		}
	}
}

func newServeMux(response string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(response))
	})
	return mux
}

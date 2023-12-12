package violet_test

import (
	"fmt"
	"github.com/CharLemAznable/ge"
	"github.com/CharLemAznable/violet"
	"github.com/CharLemAznable/violet/internal/elf"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDataPlane(t *testing.T) {
	// backends
	aServe := http.NewServeMux()
	aServe.HandleFunc("/AAA/aaa", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("AAA"))
	})
	aServer := httptest.NewServer(aServe)
	bServe := http.NewServeMux()
	bServe.HandleFunc("/BBB/bbb", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("BBB"))
	})
	bServer := httptest.NewServer(bServe)

	// config
	configFmt := `
[[Endpoint]]
  Name = "a"
  Location = "/AAA/"
  TargetURL = "%s"
[[Endpoint]]
  Name = "b"
  Location = "/ppp/"
  StripLocationPrefix = "true"
  TargetURL = "%s"
  [Endpoint.Resilience]
    [Endpoint.Resilience.Bulkhead]
      Disabled = "true"
    [Endpoint.Resilience.TimeLimiter]
      Disabled = "true"
    [Endpoint.Resilience.RateLimiter]
      Disabled = "true"
    [Endpoint.Resilience.CircuitBreaker]
      Disabled = "true"
    [Endpoint.Resilience.Retry]
      Disabled = "true"
    [Endpoint.Resilience.Cache]
      Enabled = "true"
    [Endpoint.Resilience.Fallback]
      Enabled = "true"
      FallbackResponse = "HTTP/1.1 200 OK\r\n\r\n"
      [Endpoint.Resilience.Fallback.ResponseFailedPredicateContext]
        "some_key" = "some_value"
`
	configData := fmt.Sprintf(configFmt, aServer.URL, bServer.URL+"/BBB")
	config, _ := violet.LoadConfig(configData)

	// data plane
	dataPlane := violet.NewDataPlane(config)
	if dataPlane.GetConfig() != config {
		t.Error("Expected RawConfig is same, but not")
	}
	dataPlane.SetConfig(config)
	if dataPlane.GetConfig() != config {
		t.Error("Expected RawConfig is same, but not")
	}

	server := httptest.NewServer(dataPlane)

	request, _ := http.NewRequest("GET", server.URL+"/AAA/aaa", nil)
	requestBody, _ := elf.DumpRequestBody(request)
	if requestBody != nil {
		t.Errorf("Expected requestBody is nil, but got '%s'", string(requestBody))
	}
	response, _ := server.Client().Do(request)
	responseBody, _ := elf.DumpResponseBody(response)
	if string(responseBody) != "AAA" {
		t.Errorf("Expected responseBody is 'AAA', but got '%s'", string(responseBody))
	}

	request, _ = http.NewRequest("GET", server.URL+"/ppp/bbb", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	if requestBody != nil {
		t.Errorf("Expected requestBody is nil, but got '%s'", string(requestBody))
	}
	response, _ = server.Client().Do(request)
	responseBody, _ = elf.DumpResponseBody(response)
	if string(responseBody) != "BBB" {
		t.Errorf("Expected responseBody is 'BBB', but got '%s'", string(responseBody))
	}

	if !ge.EqualsPointer(prometheus.DefaultRegisterer, dataPlane.GetRegisterer()) {
		t.Error("Expected Registerer is DefaultRegisterer, but not")
	}
	if !ge.EqualsPointer(prometheus.DefaultGatherer, dataPlane.GetGatherer()) {
		t.Error("Expected Registerer is DefaultRegisterer, but not")
	}
	newRegistry := prometheus.NewRegistry()
	dataPlane.SetRegistry(newRegistry)
	if !ge.EqualsPointer(prometheus.Registerer(newRegistry), dataPlane.GetRegisterer()) {
		t.Error("Expected Registerer is newRegistry, but not")
	}
	if !ge.EqualsPointer(prometheus.Gatherer(newRegistry), dataPlane.GetGatherer()) {
		t.Error("Expected Registerer is newRegistry, but not")
	}

	ctrlServer := httptest.NewServer(violet.NewCtrlPlane(dataPlane))

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/config", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	responseBody, _ = elf.DumpResponseBody(response)
	t.Logf("config response:\n%s", string(responseBody))

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/metrics", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	responseBody, _ = elf.DumpResponseBody(response)
	t.Logf("metrics response:\n%s", string(responseBody))

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/disable?endpoint=a", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response StatusCode is '200', but got '%d'", response.StatusCode)
	}

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/state?endpoint=a", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	responseBody, _ = elf.DumpResponseBody(response)
	if string(responseBody) != "DISABLED" {
		t.Errorf("Expected responseBody is 'DISABLED', but got '%s'", string(responseBody))
	}

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/force-open?endpoint=a", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response StatusCode is '200', but got '%d'", response.StatusCode)
	}

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/state?endpoint=a", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	responseBody, _ = elf.DumpResponseBody(response)
	if string(responseBody) != "FORCED_OPEN" {
		t.Errorf("Expected responseBody is 'FORCED_OPEN', but got '%s'", string(responseBody))
	}

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/close?endpoint=a", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response StatusCode is '200', but got '%d'", response.StatusCode)
	}

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/state?endpoint=a", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	responseBody, _ = elf.DumpResponseBody(response)
	if string(responseBody) != "CLOSED" {
		t.Errorf("Expected responseBody is 'CLOSED', but got '%s'", string(responseBody))
	}

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/disable?endpoint=b", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response StatusCode is '200', but got '%d'", response.StatusCode)
	}

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/state?endpoint=b", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	responseBody, _ = elf.DumpResponseBody(response)
	if string(responseBody) != "UNKNOWN" {
		t.Errorf("Expected responseBody is 'UNKNOWN', but got '%s'", string(responseBody))
	}

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/force-open?endpoint=b", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response StatusCode is '200', but got '%d'", response.StatusCode)
	}

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/state?endpoint=b", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	responseBody, _ = elf.DumpResponseBody(response)
	if string(responseBody) != "UNKNOWN" {
		t.Errorf("Expected responseBody is 'UNKNOWN', but got '%s'", string(responseBody))
	}

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/close?endpoint=b", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected response StatusCode is '200', but got '%d'", response.StatusCode)
	}

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/state?endpoint=b", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	responseBody, _ = elf.DumpResponseBody(response)
	if string(responseBody) != "UNKNOWN" {
		t.Errorf("Expected responseBody is 'UNKNOWN', but got '%s'", string(responseBody))
	}

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/disable?endpoint=c", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected response StatusCode is '400', but got '%d'", response.StatusCode)
	}

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/state?endpoint=c", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	responseBody, _ = elf.DumpResponseBody(response)
	if string(responseBody) != "UNKNOWN" {
		t.Errorf("Expected responseBody is 'UNKNOWN', but got '%s'", string(responseBody))
	}

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/force-open?endpoint=c", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected response StatusCode is '400', but got '%d'", response.StatusCode)
	}

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/state?endpoint=c", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	responseBody, _ = elf.DumpResponseBody(response)
	if string(responseBody) != "UNKNOWN" {
		t.Errorf("Expected responseBody is 'UNKNOWN', but got '%s'", string(responseBody))
	}

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/close?endpoint=c", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected response StatusCode is '400', but got '%d'", response.StatusCode)
	}

	request, _ = http.NewRequest("GET", ctrlServer.URL+"/circuitbreaker/state?endpoint=c", nil)
	requestBody, _ = elf.DumpRequestBody(request)
	response, _ = server.Client().Do(request)
	responseBody, _ = elf.DumpResponseBody(response)
	if string(responseBody) != "UNKNOWN" {
		t.Errorf("Expected responseBody is 'UNKNOWN', but got '%s'", string(responseBody))
	}
}

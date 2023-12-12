package violet

import (
	"bytes"
	"github.com/BurntSushi/toml"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func ConfigHandler(dataPlane DataPlane) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		buffer := new(bytes.Buffer)
		encoder := toml.NewEncoder(buffer)
		_ = encoder.Encode(dataPlane.GetConfig())
		writer.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		_, _ = writer.Write([]byte(buffer.String()))
	})
}

func MetricsHandler(dataPlane DataPlane) http.Handler {
	return promhttp.HandlerFor(dataPlane.GetGatherer(), promhttp.HandlerOpts{})
}

func DisableCircuitBreakerHandler(dataPlane DataPlane) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		name := request.FormValue("endpoint")
		if err := dataPlane.DisableCircuitBreaker(name); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

func ForceOpenCircuitBreakerHandler(dataPlane DataPlane) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		name := request.FormValue("endpoint")
		if err := dataPlane.ForceOpenCircuitBreaker(name); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

func CloseCircuitBreakerHandler(dataPlane DataPlane) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		name := request.FormValue("endpoint")
		if err := dataPlane.CloseCircuitBreaker(name); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

func CircuitBreakerStateHandler(dataPlane DataPlane) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		name := request.FormValue("endpoint")
		_, _ = writer.Write([]byte(dataPlane.CircuitBreakerState(name)))
	})
}

package violet

import "net/http"

func NewCtrlPlane(dataPlane DataPlane) http.Handler {
	ctrlPlane := new(http.ServeMux)
	ctrlPlane.Handle("/config", ConfigHandler(dataPlane))
	ctrlPlane.Handle("/metrics", MetricsHandler(dataPlane))
	ctrlPlane.Handle("/circuitbreaker/disable", DisableCircuitBreakerHandler(dataPlane))
	ctrlPlane.Handle("/circuitbreaker/force-open", ForceOpenCircuitBreakerHandler(dataPlane))
	ctrlPlane.Handle("/circuitbreaker/close", CloseCircuitBreakerHandler(dataPlane))
	ctrlPlane.Handle("/circuitbreaker/state", CircuitBreakerStateHandler(dataPlane))
	return ctrlPlane
}

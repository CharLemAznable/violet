package violet

import "github.com/CharLemAznable/violet/internal/resilience"

//goland:noinspection GoUnusedGlobalVariable
var (
	RegisterRspFailedPredicate = resilience.RegisterRspFailedPredicate
	RegisterRspCachePredicate  = resilience.RegisterRspCachePredicate
	RegisterFallbackFunction   = resilience.RegisterFallbackFunction
)

package resilience

import (
	"github.com/CharLemAznable/ge"
	"github.com/CharLemAznable/resilience4go/decorator"
	. "github.com/CharLemAznable/violet/internal/types"
)

type FallbackConfig struct {
	Enabled                        string
	FallbackResponse               string
	FallbackFunction               string
	FallbackFunctionContext        map[string]string
	ResponseFailedPredicate        string
	ResponseFailedPredicateContext map[string]string
	Order                          string
}

const FallbackDefaultOrder = "700"

func NewFallbackPlugin(config *FallbackConfig) *OrderedDecorator {
	if !ge.ToBool(config.Enabled) {
		return newOrderedDecorator(ReverseProxyIdentity, config.Order, FallbackDefaultOrder)
	}
	fallbackFn := buildFallbackFunction(
		GetFallbackFunction(config.FallbackFunction),
		config.FallbackFunctionContext)
	if fallbackFn == nil {
		fallbackFn = responseFn(config.FallbackResponse)
		if fallbackFn == nil {
			return newOrderedDecorator(ReverseProxyIdentity, config.Order, FallbackDefaultOrder)
		}
	}
	predicate := GetRspFailedPredicate(config.ResponseFailedPredicate)
	context := config.ResponseFailedPredicateContext
	predicateFn := buildFallbackResponsePredicate(predicate, context)
	return newOrderedDecorator(func(rp ReverseProxy) ReverseProxy {
		decorate := decorator.OfFunction(rp.Transport.RoundTrip).WithFallback(fallbackFn,
			func(_ Req, rsp Rsp, err error, panic any) bool { return predicateFn(rsp, err, panic) })
		rp.Transport = RoundTripperFunc(decorate.Decorate())
		return rp
	}, config.Order, FallbackDefaultOrder)
}

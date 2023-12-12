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
}

func NewFallbackPlugin(config *FallbackConfig) ReverseProxyDecorator {
	if !ge.ToBool(config.Enabled) {
		return ReverseProxyIdentity
	}
	fallbackFn := buildFallbackFunction(
		GetFallbackFunction(config.FallbackFunction),
		config.FallbackFunctionContext)
	if fallbackFn == nil {
		fallbackFn = responseFn(config.FallbackResponse)
		if fallbackFn == nil {
			return ReverseProxyIdentity
		}
	}
	predicate := GetRspFailedPredicate(config.ResponseFailedPredicate)
	context := config.ResponseFailedPredicateContext
	predicateFn := buildFallbackResponsePredicate(predicate, context)
	return func(rp ReverseProxy) ReverseProxy {
		decorate := decorator.OfFunction(rp.Transport.RoundTrip).WithFallback(fallbackFn,
			func(_ Req, rsp Rsp, err error, panic any) bool { return predicateFn(rsp, err, panic) })
		rp.Transport = RoundTripperFunc(decorate.Decorate())
		return rp
	}
}

package resilience

import (
	"github.com/CharLemAznable/ge"
	"github.com/CharLemAznable/resilience4go/bulkhead"
	"github.com/CharLemAznable/resilience4go/decorator"
	. "github.com/CharLemAznable/violet/internal/types"
	"strconv"
	"time"
)

type BulkheadConfig struct {
	Disabled           string
	MaxConcurrentCalls string
	MaxWaitDuration    string
	WhenFullResponse   string
}

func NewBulkheadPlugin(name string, config *BulkheadConfig) (bulkhead.Bulkhead, ReverseProxyDecorator) {
	if ge.ToBool(config.Disabled) {
		return nil, ReverseProxyIdentity
	}
	entry := bulkhead.NewBulkhead(name+"_bulkhead",
		bulkheadConfigBuilders(config)...)
	whenFullFn := responseFn(config.WhenFullResponse)
	return entry, func(rp ReverseProxy) ReverseProxy {
		decorate := decorator.OfFunction(rp.Transport.RoundTrip).WithBulkhead(entry)
		if whenFullFn != nil {
			decorate = decorate.WhenFull(whenFullFn)
		}
		rp.Transport = RoundTripperFunc(decorate.Decorate())
		return rp
	}
}

func bulkheadConfigBuilders(config *BulkheadConfig) []bulkhead.ConfigBuilder {
	var builders []bulkhead.ConfigBuilder
	if maxConcurrentCalls, err := strconv.ParseInt(
		config.MaxConcurrentCalls, 10, 64); err == nil {
		builders = append(builders, bulkhead.
			WithMaxConcurrentCalls(maxConcurrentCalls))
	}
	if maxWaitDuration, err := time.ParseDuration(
		config.MaxWaitDuration); err == nil {
		builders = append(builders, bulkhead.
			WithMaxWaitDuration(maxWaitDuration))
	}
	return builders
}

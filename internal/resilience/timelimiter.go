package resilience

import (
	"github.com/CharLemAznable/ge"
	"github.com/CharLemAznable/resilience4go/decorator"
	"github.com/CharLemAznable/resilience4go/timelimiter"
	. "github.com/CharLemAznable/violet/internal/types"
	"time"
)

type TimeLimiterConfig struct {
	Disabled            string
	TimeoutDuration     string
	WhenTimeoutResponse string
}

func NewTimeLimiterPlugin(name string, config *TimeLimiterConfig) (timelimiter.TimeLimiter, ReverseProxyDecorator) {
	if ge.ToBool(config.Disabled) {
		return nil, ReverseProxyIdentity
	}
	entry := timelimiter.NewTimeLimiter(name+"_timelimiter",
		timelimiterConfigBuilders(config)...)
	whenTimeoutFn := responseFn(config.WhenTimeoutResponse)
	return entry, func(rp ReverseProxy) ReverseProxy {
		decorate := decorator.OfFunction(rp.Transport.RoundTrip).WithTimeLimiter(entry)
		if whenTimeoutFn != nil {
			decorate = decorate.WhenTimeout(whenTimeoutFn)
		}
		rp.Transport = RoundTripperFunc(decorate.Decorate())
		return rp
	}
}

func timelimiterConfigBuilders(config *TimeLimiterConfig) []timelimiter.ConfigBuilder {
	var builders []timelimiter.ConfigBuilder
	if timeoutDuration, err := time.ParseDuration(
		config.TimeoutDuration); err == nil {
		builders = append(builders, timelimiter.
			WithTimeoutDuration(timeoutDuration))
	} else {
		builders = append(builders, timelimiter.
			WithTimeoutDuration(time.Second*60))
	}
	return builders
}

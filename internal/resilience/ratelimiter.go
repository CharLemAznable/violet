package resilience

import (
	"github.com/CharLemAznable/ge"
	"github.com/CharLemAznable/resilience4go/decorator"
	"github.com/CharLemAznable/resilience4go/ratelimiter"
	. "github.com/CharLemAznable/violet/internal/types"
	"strconv"
	"time"
)

type RateLimiterConfig struct {
	Disabled             string
	TimeoutDuration      string
	LimitRefreshPeriod   string
	LimitForPeriod       string
	WhenOverRateResponse string
}

func NewRateLimiterPlugin(name string, config *RateLimiterConfig) (ratelimiter.RateLimiter, ReverseProxyDecorator) {
	if ge.ToBool(config.Disabled) {
		return nil, ReverseProxyIdentity
	}
	entry := ratelimiter.NewRateLimiter(name+"_ratelimiter",
		ratelimiterConfigBuilders(config)...)
	whenOverRateFn := responseFn(config.WhenOverRateResponse)
	return entry, func(rp ReverseProxy) ReverseProxy {
		decorate := decorator.OfFunction(rp.Transport.RoundTrip).WithRateLimiter(entry)
		if whenOverRateFn != nil {
			decorate = decorate.WhenOverRate(whenOverRateFn)
		}
		rp.Transport = RoundTripperFunc(decorate.Decorate())
		return rp
	}
}

func ratelimiterConfigBuilders(config *RateLimiterConfig) []ratelimiter.ConfigBuilder {
	var builders []ratelimiter.ConfigBuilder
	if timeoutDuration, err := time.ParseDuration(
		config.TimeoutDuration); err == nil {
		builders = append(builders, ratelimiter.
			WithTimeoutDuration(timeoutDuration))
	}
	if limitRefreshPeriod, err := time.ParseDuration(
		config.LimitRefreshPeriod); err == nil {
		builders = append(builders, ratelimiter.
			WithLimitRefreshPeriod(limitRefreshPeriod))
	}
	if limitForPeriod, err := strconv.ParseInt(
		config.LimitForPeriod, 10, 64); err == nil {
		builders = append(builders, ratelimiter.
			WithLimitForPeriod(limitForPeriod))
	}
	return builders
}

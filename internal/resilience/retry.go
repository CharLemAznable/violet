package resilience

import (
	"github.com/CharLemAznable/gogo/lang"
	"github.com/CharLemAznable/resilience4go/decorator"
	"github.com/CharLemAznable/resilience4go/retry"
	. "github.com/CharLemAznable/violet/internal/types"
	"strconv"
	"time"
)

type RetryConfig struct {
	Disabled                       string
	MaxAttempts                    string
	FailAfterMaxAttempts           string
	ResponseFailedPredicate        string
	ResponseFailedPredicateContext map[string]string
	WaitInterval                   string
	WhenMaxRetriesResponse         string
	Order                          string
}

const RetryDefaultOrder = "500"

func NewRetryPlugin(name string, config *RetryConfig) (retry.Retry, *OrderedDecorator) {
	if lang.ToBool(config.Disabled) {
		return nil, newOrderedDecorator(ReverseProxyIdentity, config.Order, RetryDefaultOrder)
	}
	entry := retry.NewRetry(name+"_retry",
		retryConfigBuilders(config)...)
	whenMaxRetriesFn := responseFn(config.WhenMaxRetriesResponse)
	return entry, newOrderedDecorator(func(rp ReverseProxy) ReverseProxy {
		decorate := decorator.OfFunction(rp.Transport.RoundTrip).WithRetry(entry)
		if whenMaxRetriesFn != nil {
			decorate = decorate.WhenMaxRetries(whenMaxRetriesFn)
		}
		rp.Transport = RoundTripperFunc(decorate.Decorate())
		return rp
	}, config.Order, RetryDefaultOrder)
}

func retryConfigBuilders(config *RetryConfig) []retry.ConfigBuilder {
	var builders []retry.ConfigBuilder
	if maxAttempts, err := strconv.Atoi(config.MaxAttempts); err == nil {
		builders = append(builders, retry.WithMaxAttempts(maxAttempts))
	}
	builders = append(builders, retry.
		WithFailAfterMaxAttempts(lang.ToBool(config.FailAfterMaxAttempts)))
	predicate := GetRspFailedPredicate(config.ResponseFailedPredicate)
	context := config.ResponseFailedPredicateContext
	builders = append(builders, retry.WithFailureResultPredicate(
		buildFailureResultPredicate(predicate, context)))
	if waitInterval, err := time.ParseDuration(config.WaitInterval); err == nil {
		builders = append(builders, retry.WithWaitIntervalFunction(
			func(_ int) time.Duration { return waitInterval }))
	}
	return builders
}

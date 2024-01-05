package resilience

import (
	"bufio"
	"bytes"
	"github.com/CharLemAznable/gogo/lang"
	"github.com/CharLemAznable/resilience4go/cache"
	. "github.com/CharLemAznable/violet/internal/types"
	"github.com/dgraph-io/ristretto/z"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"
)

type CacheConfig struct {
	Enabled                       string
	Capacity                      string
	ItemTTL                       string
	ResponseCachePredicate        string
	ResponseCachePredicateContext map[string]string
	Order                         string
}

const CacheDefaultOrder = "600"

func NewCachePlugin(name string, config *CacheConfig) (cache.Cache[Req, Rsp], *OrderedDecorator) {
	if !lang.ToBool(config.Enabled) {
		return nil, newOrderedDecorator(ReverseProxyIdentity, config.Order, CacheDefaultOrder)
	}
	entry := cache.NewCache[Req, Rsp](name+"_cache",
		cacheConfigBuilders(config)...).WithMarshalFn(marshalFn, unmarshalFn)
	return entry, newOrderedDecorator(func(rp ReverseProxy) ReverseProxy {
		rt := rp.Transport
		rp.Transport = RoundTripperFunc(func(req Req) (Rsp, error) {
			rsp, err := entry.GetOrLoad(req, rt.RoundTrip)
			if rsp != nil {
				rsp.Request = req
			}
			return rsp, err
		})
		return rp
	}, config.Order, CacheDefaultOrder)
}

func cacheConfigBuilders(config *CacheConfig) []cache.ConfigBuilder {
	var builders []cache.ConfigBuilder
	if capacity, err := strconv.ParseInt(
		config.Capacity, 10, 64); err == nil {
		builders = append(builders, cache.WithCapacity(capacity))
	}
	if itemTTL, err := time.ParseDuration(config.ItemTTL); err == nil {
		builders = append(builders, cache.WithItemTTL(itemTTL))
	}
	builders = append(builders, cache.WithKeyToHash(
		func(key any) (uint64, uint64) {
			req, err := lang.Cast[Req](key)
			lang.PanicIfError(err)
			dumpRequest, err := httputil.DumpRequest(req, true)
			lang.PanicIfError(err)
			return z.KeyToHash(dumpRequest)
		}))
	predicate := GetRspCachePredicate(config.ResponseCachePredicate)
	context := config.ResponseCachePredicateContext
	builders = append(builders, cache.WithCacheResultPredicate(
		buildCacheResultPredicate(predicate, context)))
	return builders
}

func marshalFn(rsp Rsp) any {
	dumpResponse, err := httputil.DumpResponse(rsp, true)
	lang.PanicIfError(err)
	return dumpResponse
}

func unmarshalFn(v any) Rsp {
	dumpResponse, err := lang.Cast[[]byte](v)
	lang.PanicIfError(err)
	rsp, err := http.ReadResponse(bufio.
		NewReader(bytes.NewReader(dumpResponse)), nil)
	lang.PanicIfError(err)
	return rsp
}

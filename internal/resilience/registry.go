package resilience

import (
	"github.com/CharLemAznable/gogo/ext"
	"github.com/CharLemAznable/gogo/lang"
	. "github.com/CharLemAznable/violet/internal/types"
	"net/http"
)

type RspFailedPredicate = func(Rsp, map[string]string) bool

func DefaultRspFailedPredicate(rsp Rsp, _ map[string]string) bool {
	return rsp.StatusCode >= http.StatusInternalServerError
}

var rspFailedPredicateRegister = ext.NewDefaultRegistry("", DefaultRspFailedPredicate)

func RegisterRspFailedPredicate(name string, predicate RspFailedPredicate) error {
	return rspFailedPredicateRegister.Register(name, predicate)
}

func GetRspFailedPredicate(name string) RspFailedPredicate {
	predicate, _ := rspFailedPredicateRegister.Get(name)
	return predicate
}

func buildFailureResultPredicate(
	predicate RspFailedPredicate,
	context map[string]string) func(any, error) bool {
	return func(ret any, err error) bool {
		if err != nil {
			return true
		}
		rsp, err := lang.Cast[Rsp](ret)
		lang.PanicIfError(err)
		return predicate(rsp, context)
	}
}

////////////////////////////////////////////////////////////////

type RspCachePredicate = func(Rsp, map[string]string) bool

func DefaultRspCachePredicate(rsp Rsp, _ map[string]string) bool {
	return rsp.StatusCode == http.StatusOK ||
		rsp.StatusCode == http.StatusMovedPermanently ||
		rsp.StatusCode == http.StatusNotFound
}

var rspCachePredicateRegister = ext.NewDefaultRegistry("", DefaultRspCachePredicate)

func RegisterRspCachePredicate(name string, predicate RspCachePredicate) error {
	return rspCachePredicateRegister.Register(name, predicate)
}

func GetRspCachePredicate(name string) RspCachePredicate {
	predicate, _ := rspCachePredicateRegister.Get(name)
	return predicate
}

func buildCacheResultPredicate(
	predicate RspCachePredicate,
	context map[string]string) func(any, error) bool {
	return func(ret any, err error) bool {
		if err != nil {
			return false
		}
		rsp, err := lang.Cast[Rsp](ret)
		lang.PanicIfError(err)
		return predicate(rsp, context)
	}
}

////////////////////////////////////////////////////////////////

type FallbackFunction = func(Req, map[string]string) (Rsp, error)

var fallbackFunctionRegister = ext.NewDefaultRegistry("", FallbackFunction(nil))

func RegisterFallbackFunction(name string, function FallbackFunction) error {
	return fallbackFunctionRegister.Register(name, function)
}

func GetFallbackFunction(name string) FallbackFunction {
	function, _ := fallbackFunctionRegister.Get(name)
	return function
}

func buildFallbackFunction(
	function FallbackFunction,
	context map[string]string) func(Req) (Rsp, error) {
	if function == nil {
		return nil
	}
	return func(req Req) (Rsp, error) {
		return function(req, context)
	}
}

func buildFallbackResponsePredicate(
	predicate RspFailedPredicate,
	context map[string]string) func(Rsp, error, any) bool {
	return func(rsp Rsp, err error, panic any) bool {
		if err != nil || panic != nil {
			return true
		}
		return predicate(rsp, context)
	}
}

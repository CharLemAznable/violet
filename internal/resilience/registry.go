package resilience

import (
	"errors"
	"github.com/CharLemAznable/ge"
	. "github.com/CharLemAznable/violet/internal/types"
	"net/http"
	"sync"
)

type RspFailedPredicate = func(Rsp, map[string]string) bool

func DefaultRspFailedPredicate(rsp Rsp, _ map[string]string) bool {
	return rsp.StatusCode >= http.StatusInternalServerError
}

type rspFailedPredicateRegistry struct {
	sync.RWMutex
	table map[string]RspFailedPredicate
}

func (r *rspFailedPredicateRegistry) register(name string, predicate RspFailedPredicate) error {
	r.Lock()
	defer r.Unlock()
	if nil == predicate {
		return errors.New("RspFailedPredicate: predicate is nil")
	}
	if name == "" {
		return errors.New("RspFailedPredicate: illegal name \"\"")
	}
	if _, exist := r.table[name]; exist {
		return errors.New("RspFailedPredicate: multiple registrations for \"" + name + "\"")
	}
	r.table[name] = predicate
	return nil
}

func (r *rspFailedPredicateRegistry) get(name string) RspFailedPredicate {
	r.RLock()
	defer r.RUnlock()
	if name == "" {
		return DefaultRspFailedPredicate
	}
	if predicate, exist := r.table[name]; exist {
		return predicate
	}
	return DefaultRspFailedPredicate
}

var rspFailedPredicateRegister = &rspFailedPredicateRegistry{table: make(map[string]RspFailedPredicate)}

func RegisterRspFailedPredicate(name string, predicate RspFailedPredicate) error {
	return rspFailedPredicateRegister.register(name, predicate)
}

func GetRspFailedPredicate(name string) RspFailedPredicate {
	return rspFailedPredicateRegister.get(name)
}

func buildFailureResultPredicate(
	predicate RspFailedPredicate,
	context map[string]string) func(any, error) bool {
	return func(ret any, err error) bool {
		if err != nil {
			return true
		}
		rsp, err := ge.Cast[Rsp](ret)
		ge.PanicIfError(err)
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

type rspCachePredicateRegistry struct {
	sync.RWMutex
	table map[string]RspCachePredicate
}

func (r *rspCachePredicateRegistry) register(name string, predicate RspCachePredicate) error {
	r.Lock()
	defer r.Unlock()
	if nil == predicate {
		return errors.New("RspCachePredicate: predicate is nil")
	}
	if name == "" {
		return errors.New("RspCachePredicate: illegal name \"\"")
	}
	if _, exist := r.table[name]; exist {
		return errors.New("RspCachePredicate: multiple registrations for \"" + name + "\"")
	}
	r.table[name] = predicate
	return nil
}

func (r *rspCachePredicateRegistry) get(name string) RspCachePredicate {
	r.RLock()
	defer r.RUnlock()
	if name == "" {
		return DefaultRspCachePredicate
	}
	if predicate, exist := r.table[name]; exist {
		return predicate
	}
	return DefaultRspCachePredicate
}

var rspCachePredicateRegister = &rspCachePredicateRegistry{table: make(map[string]RspCachePredicate)}

func RegisterRspCachePredicate(name string, predicate RspCachePredicate) error {
	return rspCachePredicateRegister.register(name, predicate)
}

func GetRspCachePredicate(name string) RspCachePredicate {
	return rspCachePredicateRegister.get(name)
}

func buildCacheResultPredicate(
	predicate RspCachePredicate,
	context map[string]string) func(any, error) bool {
	return func(ret any, err error) bool {
		if err != nil {
			return false
		}
		rsp, err := ge.Cast[Rsp](ret)
		ge.PanicIfError(err)
		return predicate(rsp, context)
	}
}

////////////////////////////////////////////////////////////////

type FallbackFunction = func(Req, map[string]string) (Rsp, error)

type fallbackFunctionRegistry struct {
	sync.RWMutex
	table map[string]FallbackFunction
}

func (r *fallbackFunctionRegistry) register(name string, function FallbackFunction) error {
	r.Lock()
	defer r.Unlock()
	if nil == function {
		return errors.New("FallbackFunction: function is nil")
	}
	if name == "" {
		return errors.New("FallbackFunction: illegal name \"\"")
	}
	if _, exist := r.table[name]; exist {
		return errors.New("FallbackFunction: multiple registrations for \"" + name + "\"")
	}
	r.table[name] = function
	return nil
}

func (r *fallbackFunctionRegistry) get(name string) FallbackFunction {
	r.RLock()
	defer r.RUnlock()
	return r.table[name]
}

var fallbackFunctionRegister = &fallbackFunctionRegistry{table: make(map[string]FallbackFunction)}

func RegisterFallbackFunction(name string, function FallbackFunction) error {
	return fallbackFunctionRegister.register(name, function)
}

func GetFallbackFunction(name string) FallbackFunction {
	return fallbackFunctionRegister.get(name)
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

package resilience_test

import (
	"github.com/CharLemAznable/ge"
	"github.com/CharLemAznable/violet/internal/resilience"
	. "github.com/CharLemAznable/violet/internal/types"
	"testing"
)

func TestRspFailedPredicateRegistry(t *testing.T) {
	err := resilience.RegisterRspFailedPredicate("", nil)
	if err == nil {
		t.Error("Expected err but got nil")
	}
	err = resilience.RegisterRspFailedPredicate("",
		func(_ Rsp, _ map[string]string) bool { return false })
	if err == nil {
		t.Error("Expected err but got nil")
	}
	err = resilience.RegisterRspFailedPredicate("test",
		func(_ Rsp, _ map[string]string) bool { return false })
	if err != nil {
		t.Errorf("Expected err is nil but got %v", err)
	}
	err = resilience.RegisterRspFailedPredicate("test",
		func(_ Rsp, _ map[string]string) bool { return false })
	if err == nil {
		t.Error("Expected err but got nil")
	}

	predicate := resilience.GetRspFailedPredicate("")
	if !ge.EqualsPointer(resilience.DefaultRspFailedPredicate, predicate) {
		t.Error("Expected get DefaultRspFailedPredicate but not")
	}
	predicate = resilience.GetRspFailedPredicate("test")
	if predicate == nil {
		t.Error("Expected get predicate but got nil")
	}
	predicate = resilience.GetRspFailedPredicate("not exist")
	if !ge.EqualsPointer(resilience.DefaultRspFailedPredicate, predicate) {
		t.Error("Expected get DefaultRspFailedPredicate but not")
	}
}

func TestRspCachePredicateRegistry(t *testing.T) {
	err := resilience.RegisterRspCachePredicate("", nil)
	if err == nil {
		t.Error("Expected err but got nil")
	}
	err = resilience.RegisterRspCachePredicate("",
		func(_ Rsp, _ map[string]string) bool { return true })
	if err == nil {
		t.Error("Expected err but got nil")
	}
	err = resilience.RegisterRspCachePredicate("test",
		func(_ Rsp, _ map[string]string) bool { return true })
	if err != nil {
		t.Errorf("Expected err is nil but got %v", err)
	}
	err = resilience.RegisterRspCachePredicate("test",
		func(_ Rsp, _ map[string]string) bool { return true })
	if err == nil {
		t.Error("Expected err but got nil")
	}

	predicate := resilience.GetRspCachePredicate("")
	if !ge.EqualsPointer(resilience.DefaultRspCachePredicate, predicate) {
		t.Error("Expected get DefaultRspCachePredicate but not")
	}
	predicate = resilience.GetRspCachePredicate("test")
	if predicate == nil {
		t.Error("Expected get predicate but got nil")
	}
	predicate = resilience.GetRspCachePredicate("not exist")
	if !ge.EqualsPointer(resilience.DefaultRspCachePredicate, predicate) {
		t.Error("Expected get DefaultRspCachePredicate but not")
	}
}

func TestFallbackFunctionRegistry(t *testing.T) {
	err := resilience.RegisterFallbackFunction("", nil)
	if err == nil {
		t.Error("Expected err but got nil")
	}
	err = resilience.RegisterFallbackFunction("",
		func(_ Req, _ map[string]string) (Rsp, error) { return nil, nil })
	if err == nil {
		t.Error("Expected err but got nil")
	}
	err = resilience.RegisterFallbackFunction("test",
		func(_ Req, _ map[string]string) (Rsp, error) { return nil, nil })
	if err != nil {
		t.Errorf("Expected err is nil but got %v", err)
	}
	err = resilience.RegisterFallbackFunction("test",
		func(_ Req, _ map[string]string) (Rsp, error) { return nil, nil })
	if err == nil {
		t.Error("Expected err but got nil")
	}

	function := resilience.GetFallbackFunction("")
	if function != nil {
		t.Errorf("Expected get nil function but got %v", any(function))
	}
	function = resilience.GetFallbackFunction("test")
	if function == nil {
		t.Error("Expected get predicate but got nil")
	}
	function = resilience.GetFallbackFunction("not exist")
	if function != nil {
		t.Errorf("Expected get nil function but got %v", any(function))
	}
}

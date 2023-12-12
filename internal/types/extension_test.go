package types_test

import (
	"errors"
	"github.com/CharLemAznable/ge"
	"github.com/CharLemAznable/violet/internal/types"
	"net/http/httputil"
	"testing"
)

func TestReverseProxyIdentity(t *testing.T) {
	reverseProxy := &httputil.ReverseProxy{}
	decorated := types.ReverseProxyIdentity(reverseProxy)
	if !ge.EqualsPointer(reverseProxy, decorated) {
		t.Error("Expected equal pointer but not")
	}
}

func TestRoundTripperFunc(t *testing.T) {
	fn := types.RoundTripperFunc(func(req types.Req) (types.Rsp, error) {
		return nil, errors.New("error")
	})
	_, err := fn.RoundTrip(nil)
	if "error" != err.Error() {
		t.Errorf("Expected err is 'error', but got '%s'", err.Error())
	}
}

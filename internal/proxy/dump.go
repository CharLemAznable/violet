package proxy

import (
	"github.com/CharLemAznable/gogo/ext"
	. "github.com/CharLemAznable/violet/internal/types"
	"net/http/httputil"
	"time"
)

type DumpType string

const (
	TargetDump DumpType = "TARGET"
	SourceDump DumpType = "SOURCE"
)

type MessageType string

const (
	ReqMessage MessageType = "REQ"
	RspMessage MessageType = "RSP"
)

type DumpMessage struct {
	DumpType    DumpType
	Name        string
	MessageType MessageType
	UnixNano    int64
	Content     []byte
	Error       error
}

const DumpTopic = "violet.proxy.dump"

func DumpDecorator(dump bool, dumpType DumpType, name string) ReverseProxyDecorator {
	if !dump {
		return ReverseProxyIdentity
	}
	return func(rp ReverseProxy) ReverseProxy {
		rt := rp.Transport
		rp.Transport = RoundTripperFunc(func(request Req) (Rsp, error) {
			publishRequest(request, dumpType, name)
			response, err := rt.RoundTrip(request)
			publishResponse(response, dumpType, name)
			return response, err
		})
		return rp
	}
}

func publishRequest(request Req, dumpType DumpType, name string) {
	if request != nil {
		message := &DumpMessage{
			DumpType:    dumpType,
			Name:        name,
			MessageType: ReqMessage,
			UnixNano:    time.Now().UnixNano(),
		}
		message.Content, message.Error = httputil.DumpRequest(request, true)
		go ext.Publish(DumpTopic, message)
	}
}

func publishResponse(response Rsp, dumpType DumpType, name string) {
	if response != nil {
		message := &DumpMessage{
			DumpType:    dumpType,
			Name:        name,
			MessageType: RspMessage,
			UnixNano:    time.Now().UnixNano(),
		}
		message.Content, message.Error = httputil.DumpResponse(response, true)
		go ext.Publish(DumpTopic, message)
	}
}

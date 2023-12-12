package types

func ReverseProxyIdentity(rp ReverseProxy) ReverseProxy {
	return rp
}

type RoundTripperFunc func(Req) (Rsp, error)

func (fn RoundTripperFunc) RoundTrip(r Req) (Rsp, error) {
	return fn(r)
}

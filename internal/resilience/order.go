package resilience

import . "github.com/CharLemAznable/violet/internal/types"

type OrderedDecorator struct {
	Decorator ReverseProxyDecorator
	order     string
}

func (d *OrderedDecorator) Decorate(rp ReverseProxy) ReverseProxy {
	return d.Decorator(rp)
}

func (d *OrderedDecorator) Order() string {
	return d.order
}

func newOrderedDecorator(
	decorator ReverseProxyDecorator,
	order string, defaultOrder string) *OrderedDecorator {
	if order == "" {
		return &OrderedDecorator{Decorator: decorator, order: order}
	}
	return &OrderedDecorator{Decorator: decorator, order: defaultOrder}
}

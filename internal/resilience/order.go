package resilience

import (
	. "github.com/CharLemAznable/violet/internal/types"
	"sort"
)

type OrderedDecorator struct {
	Decorator ReverseProxyDecorator
	Order     string
}

func (d *OrderedDecorator) Decorate(rp ReverseProxy) ReverseProxy {
	return d.Decorator(rp)
}

type OrderedDecoratorSlice []*OrderedDecorator

func (x OrderedDecoratorSlice) Len() int           { return len(x) }
func (x OrderedDecoratorSlice) Less(i, j int) bool { return x[i].Order < x[j].Order }
func (x OrderedDecoratorSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

func (x OrderedDecoratorSlice) Sort() { sort.Sort(x) }

func newOrderedDecorator(
	decorator ReverseProxyDecorator,
	order string, defaultOrder string) *OrderedDecorator {
	if order == "" {
		return &OrderedDecorator{Decorator: decorator, Order: order}
	}
	return &OrderedDecorator{Decorator: decorator, Order: defaultOrder}
}

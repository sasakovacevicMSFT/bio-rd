package actions

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type ASPathPrependAction struct {
	asn   uint32
	times uint16
}

func NewASPathPrependAction(asn uint32, times uint16) *ASPathPrependAction {
	return &ASPathPrependAction{
		asn:   asn,
		times: times,
	}
}

func (a *ASPathPrependAction) Do(p net.Prefix, pa *route.Path) Result {
	if pa.BGPPath == nil {
		return Result{Path: pa}
	}

	modified := pa.Copy()
	modified.BGPPath.Prepend(a.asn, a.times)

	return Result{Path: modified}
}

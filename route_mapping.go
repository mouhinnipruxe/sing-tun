package tun

import (
	"net/netip"
	"time"

	"github.com/metacubex/sing/common/cache"
)

type DirectRouteSession struct {
	// IPVersion uint8
	// Network     uint8
	Source      netip.Addr
	Destination netip.Addr
}

type RouteMapping struct {
	status *cache.LruCache[DirectRouteSession, DirectRouteDestination]
}

func NewRouteMapping(timeout time.Duration) *RouteMapping {
	status := cache.New[DirectRouteSession, DirectRouteDestination](
		cache.WithSize[DirectRouteSession, DirectRouteDestination](1024),
		cache.WithEvict[DirectRouteSession, DirectRouteDestination](func(session DirectRouteSession, action DirectRouteDestination) {
			action.Close()
		}),
		cache.WithUpdateAgeOnGet[DirectRouteSession, DirectRouteDestination](),
		cache.WithAge[DirectRouteSession, DirectRouteDestination](int64(timeout.Seconds())),
	)
	return &RouteMapping{status}
}

func (m *RouteMapping) Lookup(session DirectRouteSession, constructor func() (DirectRouteDestination, error)) (DirectRouteDestination, error) {
	var (
		created DirectRouteDestination
		err     error
	)
	action, _, ok := m.status.LoadOrStoreEx(session, func() (DirectRouteDestination, bool) {
		created, err = constructor()
		return created, err == nil
	})
	if !ok {
		return nil, err
	}
	return action, nil
}

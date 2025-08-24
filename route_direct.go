package tun

import (
	"net/netip"
	"time"

	"github.com/metacubex/sing/common/buf"
	"github.com/metacubex/sing/common/cache"
)

type DirectRouteDestination interface {
	WritePacket(packet *buf.Buffer) error
	Close() error
	IsClosed() bool
}

type DirectRouteSession struct {
	// IPVersion uint8
	// Network     uint8
	Source      netip.Addr
	Destination netip.Addr
}

type DirectRouteMapping struct {
	status *cache.LruCache[DirectRouteSession, DirectRouteDestination]
}

func NewDirectRouteMapping(timeout time.Duration) *DirectRouteMapping {
	//mapping.SetHealthCheck(func(session DirectRouteSession, destination DirectRouteDestination) bool {
	//	return !destination.IsClosed()
	//})
	status := cache.New[DirectRouteSession, DirectRouteDestination](
		cache.WithSize[DirectRouteSession, DirectRouteDestination](1024),
		cache.WithEvict[DirectRouteSession, DirectRouteDestination](func(session DirectRouteSession, action DirectRouteDestination) {
			action.Close()
		}),
		cache.WithUpdateAgeOnGet[DirectRouteSession, DirectRouteDestination](),
		cache.WithAge[DirectRouteSession, DirectRouteDestination](int64(timeout.Seconds())),
	)
	return &DirectRouteMapping{status}
}

func (m *DirectRouteMapping) Lookup(session DirectRouteSession, constructor func() (DirectRouteDestination, error)) (DirectRouteDestination, error) {
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

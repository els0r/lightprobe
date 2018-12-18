package capture

import (
	"errors"

	"github.com/els0r/lightprobe/pkg/flow"
)

// ConntrackPacketFetcher implements the PacketFetcher interface for packets retrieved via conntrack
type ConntrackPacketFetcher struct{}

// Open opens the connection to conntrack
func (c *ConntrackPacketFetcher) Open() error {
	return errors.New("Conntrack packet fetching is only available on linux platforms")
}

// Fetch fetches packets from conntrack via a list of listeneres. Currently, it listens on New and Update events
func (c *ConntrackPacketFetcher) Fetch() <-chan *flow.Packet {
	// generate packet channel and conntrack hook
	packets := make(chan *flow.Packet)
	return packets
}

// Close closes all running listeners and closes the connection to conntrack
func (c *ConntrackPacketFetcher) Close() {}

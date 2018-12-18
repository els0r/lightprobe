package capture

import (
	"github.com/els0r/lightprobe/pkg/flow"
)

// MockPacketFetcher implements the PacketFetcher interface and creates packet objects out of thin air. It's only purpose is for testing
type MockPacketFetcher struct{}

func (m *MockPacketFetcher) Open() error {
	return nil
}

func (m *MockPacketFetcher) Fetch() <-chan *flow.Packet {
	packets := make(chan *flow.Packet)

	var direction bool
	go func() {
		for j := 0; j < 2; j++ {
			for i := 0; i < 10; i++ {
				p := flow.NewPacket()
				p.Attributes.SIP[15] = 0x01
				p.Attributes.DIP[15] = 0x01
				p.Attributes.Sport[0] = byte(i)
				p.Attributes.Dport[0] = byte(15 - i)
				p.Attributes.IPProtocol = byte(i)
				p.NBytes = 10
				p.Iface = "eth0"
				p.Origin = direction

				packets <- p
			}
			direction = true
		}
	}()

	return packets
}

func (m *MockPacketFetcher) Close() {
}

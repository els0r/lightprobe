package capture

// functional options for the capture object
func WithPacketSource(source PacketFetcher) func(*Capture) {
	return func(c *Capture) {
		c.source = source
	}
}

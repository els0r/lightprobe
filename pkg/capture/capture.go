package capture

import (
	"errors"
	"fmt"
	"time"

	"github.com/els0r/lightprobe/pkg/config"
	"github.com/els0r/lightprobe/pkg/flow"
)

type Capture struct {
	quitChan   chan bool
	source     PacketFetcher
	numPackets uint64 // counter for number of packets captured
}

type PacketFetcher interface {
	Open() error
	Fetch() <-chan *flow.Packet
	Close()
}

func New(cfg *config.Config, options ...func(*Capture)) *Capture {

	c := new(Capture)
	c.quitChan = make(chan bool)

	// Execute functional options (if any), see options.go for implementation
	for _, option := range options {
		option(c)
	}

	return c
}

func (c *Capture) Open() error {
	if c.source == nil {
		return errors.New("No packet source defined")
	}

	return c.source.Open()
}

func (c *Capture) Run(writeChan chan flow.Map, flushChan <-chan time.Time) {
	go func() {

		// get packet channel
		packets := c.source.Fetch()
		defer c.source.Close()

		flows := flow.NewMap()

		for {
			select {
			case <-c.quitChan:
				fmt.Printf("Received quit signal. Stopping capture. Final flow map: %s\n", flows)

				return
			case <-flushChan:
				// rotation handling
				if flows != nil {
				} else {

				}
			case p := <-packets:
				c.numPackets++

				// handle packet
				fmt.Printf("Received packet %d: %s\n", c.numPackets, p)
				flows.Update(p)

			default:
				fmt.Printf("Waiting for packets ...\n")
				time.Sleep(1 * time.Second)
			}
		}
	}()
}

func (c *Capture) Close() {
	c.quitChan <- true
}

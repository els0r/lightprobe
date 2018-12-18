package capture

import (
	"context"
	"fmt"
	"sync"

	"github.com/els0r/lightprobe/pkg/flow"

	ct "github.com/florianl/go-conntrack"
)

// ConntrackPacketFetcher implements the PacketFetcher interface for packets retrieved via conntrack
type ConntrackPacketFetcher struct {
	quitUpdateListener chan *sync.WaitGroup
	quitNewListener    chan *sync.WaitGroup
	conn               *ct.Nfct //conntrack object
}

type ctListener struct {
	onEvent ct.NetlinkGroup
	quit    chan *sync.WaitGroup
}

// NewConntrackPacketFetcher returns a new conntrack fetcher
func NewConntrackPacketFetcher() *ConntrackPacketFetcher {
	return &ConntrackPacketFetcher{
		quitUpdateListener: make(chan *sync.WaitGroup),
		quitNewListener:    make(chan *sync.WaitGroup),
	}
}

// Open opens the connection to conntrack
func (c *ConntrackPacketFetcher) Open() error {
	var err error

	// open the conntrack link
	c.conn, err = ct.Open()
	return err
}

// Fetch fetches packets from conntrack via a list of listeneres. Currently, it listens on New and Update events
func (c *ConntrackPacketFetcher) Fetch() <-chan *flow.Packet {

	// generate packet channel and conntrack hook
	packets := make(chan *flow.Packet)

	hook := generateCTPacketHandler(packets)

	// spawn listeners
	for _, listener := range []ctListener{
		ctListener{ct.NetlinkCtNew, c.quitNewListener},
		ctListener{ct.NetlinkCtUpdate, c.quitUpdateListener},
	} {
		go func() {
			// register routines
			errChan, err := c.conn.Register(
				context.Background(),
				ct.Ct, listener.onEvent,
				hook,
			)

			// TODO: check how to properly handle this
			if err != nil {
				fmt.Printf("ConntrackPacketFetcher: could not register listener [nl=%d]: %s\n", listener.onEvent, err)
				return
			}

			for {
				select {
				case err := <-errChan:
					fmt.Printf("ConntrackPacketFetcher: [nl=%d]: Error: %s\n", listener.onEvent, err)
				case wg := <-listener.quit:
					wg.Add(1)
					fmt.Printf("ConntrackPacketFetcher: [nl=%d]: Closing down...\n", listener.onEvent)
					wg.Done()
					return
				}
			}
		}()
	}

	return packets
}

// Close closes all running listeners and closes the connection to conntrack
func (c *ConntrackPacketFetcher) Close() {

	// don't push updates on quit channel if nothing is running
	if c.conn == nil {
		return
	}

	var wg sync.WaitGroup

	c.quitUpdateListener <- &wg
	c.quitNewListener <- &wg

	wg.Wait()

	c.conn.Close()
}
func generateCTPacketHandler(packets chan *flow.Packet) func(ct.Conn) int {
	return func(c ct.Conn) int {
		var err error
		p := new(flow.Packet)

		err = populatePacketFromCTConn(c, p)
		if err != nil {
			return 1
		}

		packets <- p
		return 0
	}
}

func populatePacketFromCTConn(c ct.Conn, p *flow.Packet) error {
	var err error
	defer func() {
		if err != nil {
			fmt.Println(err)
		}
	}()

	err = getSIP(c, &p.Attributes)
	if err != nil {
		return err
	}
	err = getDIP(c, &p.Attributes)
	if err != nil {
		return err
	}

	getSport(c, &p.Attributes)
	getDport(c, &p.Attributes)
	getIPProtocol(c, &p.Attributes)

	return err
}

func getSIP(c ct.Conn, p *flow.PacketAttributeBytes) error {
	if data, ok := c[ct.AttrOrigIPv6Src]; ok {
		copy(p.SIP[:], data[:16])
		return nil
	} else if data, ok := c[ct.AttrOrigIPv4Src]; ok {
		copy(p.SIP[:4], data[:4])
		return nil
	}
	return ct.ErrConnNoSrcIP
}

func getDIP(c ct.Conn, p *flow.PacketAttributeBytes) error {
	if data, ok := c[ct.AttrOrigIPv6Dst]; ok {
		copy(p.DIP[:], data[:16])
		return nil
	} else if data, ok := c[ct.AttrOrigIPv4Dst]; ok {
		copy(p.DIP[:4], data[:4])
		return nil
	}
	return ct.ErrConnNoDstIP
}

func getSport(c ct.Conn, p *flow.PacketAttributeBytes) {
	if data, ok := c[ct.AttrOrigPortSrc]; ok {
		copy(p.Sport[:], data[:2])
	}
}

func getDport(c ct.Conn, p *flow.PacketAttributeBytes) {
	if data, ok := c[ct.AttrOrigPortDst]; ok {
		copy(p.Dport[:], data[:2])
	}
}

func getIPProtocol(c ct.Conn, p *flow.PacketAttributeBytes) {
	if data, ok := c[ct.AttrOrigL4Proto]; ok {
		p.IPProtocol = data[0]
	}
}

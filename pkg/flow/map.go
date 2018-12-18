package flow

import (
	"fmt"
	"sync"

	"github.com/mdlayher/netlink/nlenc"
)

var (
	IPZeroBytes = [16]byte{
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}
	PortZeroBytes       = [2]byte{0x00, 0x00}
	IPProtocolZeroBytes = 0x00
)

type Packet struct {
	// base attributes
	Attributes PacketAttributeBytes

	// metadata of packet
	Iface  string // on which interface was this packet observed
	NBytes uint16 // packets cannot get larger than 2^16 bytes
	Origin bool   // direction of packet
}

func NewPacket() *Packet {
	return &Packet{
		Attributes: PacketAttributeBytes{},
	}
}

func (p *Packet) String() string {
	return fmt.Sprintf("iface=%s; len=%d; %s", p.Iface, p.NBytes, p.Attributes)
}

type PacketAttributeBytes struct {
	SIP, DIP     [16]byte // both IPv4 and IPv6 are stored within 16 byte
	Sport, Dport [2]byte
	IPProtocol   byte
}

func (p PacketAttributeBytes) String() string {
	sip, dip := rawIPToString(p.SIP[:]), rawIPToString(p.DIP[:])

	// we have to reverse the arrays due to the endianness
	sport, dport := nlenc.Uint16([]byte{p.Sport[1], p.Sport[0]}), nlenc.Uint16([]byte{p.Dport[1], p.Dport[0]})
	return fmt.Sprintf("proto=%d; %s.%d -> %s.%d", p.IPProtocol, sip, sport, dip, dport)
}

type Counters struct {
	BytesReceived, BytesSent, PacketsReceived, PacketsSent uint64
}

func (c *Counters) String() string {
	return fmt.Sprintf("Bytes: RX=%d TX=%d; Packets: RX=%d TX=%d",
		c.BytesReceived, c.BytesSent,
		c.PacketsReceived, c.PacketsSent,
	)
}

type ifaceFlows map[PacketAttributeBytes]*Counters

type Map struct {
	sync.Mutex
	Flows map[string]ifaceFlows
}

func NewMap() *Map {
	return &Map{
		Flows: make(map[string]ifaceFlows),
	}
}

func (m *Map) Update(pack *Packet) {
	m.Lock()
	defer m.Unlock()

	// check if interface is tracked already
	var (
		ifmap    ifaceFlows
		counters *Counters
		exists   bool
	)

	ifmap, exists = m.Flows[pack.Iface]
	if !exists {
		ifmap = make(map[PacketAttributeBytes]*Counters)
		m.Flows[pack.Iface] = ifmap
	}

	// initialize flow if it doesn't exist yet
	counters, exists = ifmap[pack.Attributes]
	if !exists {
		counters = &Counters{}
		ifmap[pack.Attributes] = counters
	}

	// update counters
	if pack.Origin {
		counters.BytesSent += uint64(pack.NBytes)
		counters.PacketsSent++
	} else {
		counters.BytesReceived += uint64(pack.NBytes)
		counters.PacketsReceived++
	}
}

func (m *Map) String() string {
	var str string

	for iface, flows := range m.Flows {
		str += fmt.Sprintf("%s:\n", iface)
		for pattr, counters := range flows {
			str += fmt.Sprintf("\t%s; %s\n", pattr, counters)
		}
		str += "\n"
	}

	return str
}

// GOOGLE's utility functions for printing IPv4/6 addresses ----------------------
// Convert i to hexadecimal string
func itox(i uint, min int) string {

	// Assemble hexadecimal in reverse order.
	var b [32]byte
	bp := len(b)
	for ; i > 0 || min > 0; i /= 16 {
		bp--
		b[bp] = "0123456789abcdef"[byte(i%16)]
		min--
	}

	return string(b[bp:])
}

// Convert i to decimal string.
func itod(i uint) string {
	if i == 0 {
		return "0"
	}

	// Assemble decimal in reverse order.
	var b [32]byte
	bp := len(b)
	for ; i > 0; i /= 10 {
		bp--
		b[bp] = byte(i%10) + '0'
	}

	return string(b[bp:])
}

/// END GOOGLE ///

// convert the ip byte arrays to string. The formatting logic for IPv6
// is directly copied over from the go IP package in order to save an
// additional import just for string operations
func rawIPToString(ip []byte) string {
	var (
		numZeros uint8 = 0
		iplen    int   = len(ip)
	)

	// count zeros in order to determine whether the address
	// is IPv4 or IPv6
	for i := 4; i < iplen; i++ {
		if (ip[i] & 0xFF) == 0x00 {
			numZeros++
		}
	}

	// construct ipv4 string
	if numZeros == 12 {
		return itod(uint(ip[0])) + "." +
			itod(uint(ip[1])) + "." +
			itod(uint(ip[2])) + "." +
			itod(uint(ip[3]))
	} else {
		/// START OF GOOGLE CODE SNIPPET ///
		p := ip

		// Find longest run of zeros.
		e0 := -1
		e1 := -1
		for i := 0; i < iplen; i += 2 {
			j := i
			for j < iplen && p[j] == 0 && p[j+1] == 0 {
				j += 2
			}
			if j > i && j-i > e1-e0 {
				e0 = i
				e1 = j
			}
		}

		// The symbol "::" MUST NOT be used to shorten just one 16 bit 0 field.
		if e1-e0 <= 2 {
			e0 = -1
			e1 = -1
		}

		// Print with possible :: in place of run of zeros
		var s string
		for i := 0; i < iplen; i += 2 {
			if i == e0 {
				s += "::"
				i = e1
				if i >= iplen {
					break
				}
			} else if i > 0 {
				s += ":"
			}
			s += itox((uint(p[i])<<8)|uint(p[i+1]), 1)

		}
		return s
	}
}

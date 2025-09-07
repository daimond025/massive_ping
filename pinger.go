package ping

import (
	"golang.org/x/net/icmp"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	ProtocolICMP   = 1
	ProtocolICMPv6 = 58
)

var sequence uint32

type Pinger struct {
	LogUnexpectedPackets bool // increases log verbosity
	Id                   uint16
	SequenceCounter      *uint32

	dataload  DataLoad
	payloadMu sync.RWMutex

	requests map[uint32]simpleRequest //  running requests
	mtx      sync.RWMutex             // lock for the requests map
	conn4    *net.PacketConn
	conn6    *net.PacketConn
	connUdp4 *net.PacketConn

	target      map[string]destination
	target_cidr map[string]destination_cidr
	history     map[string]history

	write4   sync.Mutex
	write6   sync.Mutex
	stat_add sync.Mutex

	complate sync.WaitGroup

	wg      sync.WaitGroup
	wg_task sync.WaitGroup
}

func connectICMP(network, address string) (net.PacketConn, error) {
	if network == "" || address == "" {
		return nil, nil
	}
	return icmp.ListenPacket(network, address)
}

func NewPinger() (*Pinger, error) {
	pinger := Pinger{
		Id:              uint16(os.Getpid()),
		SequenceCounter: &sequence,
		requests:        make(map[uint32]simpleRequest),
		target:          make(map[string]destination),
		target_cidr:     make(map[string]destination_cidr),
		history:         make(map[string]history),
	}
	pinger.SetPayloadSize(56)

	return &pinger, nil
}

func (pinger *Pinger) CreateConnection(bind4, bind6 string) error {
	conn4, err_4 := connectICMP("ip4:icmp", bind4)
	if err_4 != nil {
		return err_4
	}
	conn6, err_6 := connectICMP("ip6:ipv6-icmp", bind6)
	if err_6 != nil {
		if conn4 != nil {
			conn4.Close()
		}
		return err_6
	}

	if conn4 == nil && conn6 == nil {
		return errNotBound
	}

	pinger.conn4 = &conn4
	pinger.conn6 = &conn6
	pinger.requests = make(map[uint32]simpleRequest)
	pinger.SetPayloadSize(56)

	if conn4 != nil {
		pinger.wg.Add(1)
		go pinger.receiver(ProtocolICMP, *pinger.conn4)
	}
	if conn6 != nil {
		pinger.wg.Add(1)
		go pinger.receiver(ProtocolICMPv6, *pinger.conn6)
	}
	return nil
}

func (pinger *Pinger) Targets(addres string) {
	addres = strings.TrimSpace(addres)
	addreses := strings.Split(addres, " ")

	for _, adr := range addreses {
		adr = strings.Replace(adr, " ", "", -1)
		adr_v4, err_v4 := net.ResolveIPAddr("ip4", adr)
		if err_v4 == nil {
			pinger.target[adr] = destination{remote: adr_v4, host: adr}
			continue
		}

		adr_v6, err_v6 := net.ResolveIPAddr("ip6", adr)
		if err_v6 == nil {
			pinger.target[adr] = destination{remote: adr_v6, host: adr}
			continue
		}
	}
}

func (pinger *Pinger) Targets_CIDR(addres string) {
	addres = strings.TrimSpace(addres)
	addreses := strings.Split(addres, " ")

	for _, adr := range addreses {
		adr = strings.Replace(adr, " ", "", -1)

		ip, ipnet, err := net.ParseCIDR(adr)

		if err == nil {
			if len(ipnet.IP) == net.IPv4len {
				pinger.target_cidr[adr] = destination_cidr{ip: ip, net: ipnet, type_net: net.IPv4len}
			}
			if len(ipnet.IP) == net.IPv6len {
				pinger.target_cidr[adr] = destination_cidr{ip: ip, net: ipnet, type_net: net.IPv6len}
			}
		}

	}
}

func (pinger *Pinger) SetPayloadSize(size uint16) {
	pinger.payloadMu.Lock()
	pinger.dataload.Resize(size)
	pinger.payloadMu.Unlock()
}
func (pinger *Pinger) removeRequest(idseq uint32) {
	pinger.mtx.Lock()
	delete(pinger.requests, idseq)
	pinger.mtx.Unlock()
}

func (pinger *Pinger) SetPayload(data []byte) {
	pinger.payloadMu.Lock()
	defer pinger.payloadMu.Unlock()
	pinger.dataload = DataLoad(data)
}

func (pinger *Pinger) PayloadSize() uint16 {
	pinger.payloadMu.RLock()
	defer pinger.payloadMu.RUnlock()
	return uint16(len(pinger.dataload))
}

func (pinger *Pinger) close(conn net.PacketConn) {
	if conn != nil {
		conn.Close()
	}
	pinger.wg.Done()
}
func (pinger *Pinger) Close() {
	pinger.close(*pinger.conn4)
	pinger.close(*pinger.conn6)

	pinger.wg.Wait()
}

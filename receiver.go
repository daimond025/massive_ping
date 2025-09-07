package main

import (
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"net"
	"time"
)

func (pinger *Pinger) receiver(proto int, conn net.PacketConn) {
	rb := make([]byte, 1500)

	for {
		if n, source, err := conn.ReadFrom(rb); err != nil {
			if netErr, ok := err.(net.Error); !ok || !netErr.Temporary() {
				break // socket gone
			}
		} else {
			pinger.receive(proto, rb[:n], source.(*net.IPAddr).IP, time.Now())
		}
	}
}

func (pinger *Pinger) receive(proto int, bytes []byte, addr net.IP, t time.Time) {
	m, err := icmp.ParseMessage(proto, bytes)
	if err != nil {
		return
	}
	switch m.Type {
	case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeExtendedEchoReply:
		pinger.process(m.Body, nil, addr, &t)

	case ipv4.ICMPTypeDestinationUnreachable, ipv6.ICMPTypeDestinationUnreachable:
		body := m.Body.(*icmp.DstUnreach)
		if body == nil {
			return
		}
		var bodyData []byte

		switch proto {
		case ProtocolICMP:
			hdr, err := ipv4.ParseHeader(body.Data)
			if err != nil {
				return
			}
			bodyData = body.Data[hdr.Len:]
		case ProtocolICMPv6:
			_, err := ipv6.ParseHeader(body.Data)
			if err != nil {
				return
			}
			bodyData = body.Data[ipv6.HeaderLen:]
		default:
			return
		}
		msg, err := icmp.ParseMessage(proto, bodyData)
		if err != nil {
			return
		}

		pinger.process(msg.Body, fmt.Errorf("%s", m.Type), addr, &t)
	}
}
func (pinger *Pinger) process(body icmp.MessageBody, result error, addr net.IP, tRecv *time.Time) {
	echo, ok := body.(*icmp.Echo)
	if !ok || echo == nil {
		if pinger.LogUnexpectedPackets {
			log.Infof("expected *icmp.Echo, got %#v", body)
		}
		return
	}
	if result != nil {
		return
	}
	idseq := (uint32(uint16(echo.ID)) << 16) | uint32(uint16(echo.Seq))

	pinger.mtx.Lock()
	req, ok_exist := pinger.requests[idseq]
	if ok_exist {
		delete(pinger.requests, idseq)
		req.handler(result, addr, tRecv)
	}
	pinger.mtx.Unlock()

}

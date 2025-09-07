package main

import (
	"context"
	_ "errors"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

func (pinger *Pinger) PingAttempts(destination *net.IPAddr, timeout time.Duration, attempts int) (rtt time.Duration, err error) {
	if attempts < 1 {
		attempts = 1
	}
	for i := 0; i < attempts; i++ {
		rtt, err = pinger.Ping(destination, timeout, i)
		if err == nil {
			break // success
		}

	}
	return rtt, err
}
func (pinger *Pinger) Ping(destination *net.IPAddr, timeout time.Duration, attempt int) (time.Duration, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(timeout))
	defer cancel()
	return pinger.PingContext(ctx, destination, attempt)
}

func (pinger *Pinger) PingContext(ctx context.Context, destination *net.IPAddr, attempt int) (time.Duration, error) {
	req := create()

	idseq, err := pinger.sendRequest(destination, req, attempt)
	if err != nil {
		return 0, err
	}

	reply := Reply{}
	select {
	case <-ctx.Done():
		pinger.removeRequest(idseq)
		req.result = &timeoutError{}
	case reply = <-req.reply:
		req.close()
		req.tFinish = reply.Recieve
	}
	return req.rrt()
}

func (pinger *Pinger) sendRequest(destination *net.IPAddr, req simpleRequest, attempt int) (uint32, error) {
	pinger.payloadMu.RLock()

	id := pinger.Id
	seq := uint16(atomic.AddUint32(pinger.SequenceCounter, uint32(attempt)+1))
	idseq := (uint32(id) << 16) | uint32(seq)

	pinger.payloadMu.RUnlock()

	// build packet
	wm := icmp.Message{
		Code: 0,
		Body: &icmp.Echo{
			ID:   int(id),
			Seq:  int(seq),
			Data: pinger.dataload,
		},
	}

	var conn net.PacketConn
	var lock *sync.Mutex
	if destination.IP.To4() != nil {
		wm.Type = ipv4.ICMPTypeEcho
		conn = *pinger.conn4
		lock = &pinger.write4
	} else {
		wm.Type = ipv6.ICMPTypeEchoRequest
		conn = *pinger.conn6
		lock = &pinger.write6
	}
	wb, err := wm.Marshal(nil)

	pinger.mtx.Lock()
	pinger.requests[idseq] = req
	pinger.mtx.Unlock()

	lock.Lock()
	req.init()

	_, err = conn.WriteTo(wb, destination)
	lock.Unlock()

	// send failed, need to remove request from list
	if err != nil {
		req.close()
		pinger.removeRequest(idseq)
		return idseq, err
	}

	return idseq, nil
}

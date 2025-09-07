package main

import (
	"net"
	"sync"
	"time"
)

type request interface {
	init()
	close()
	handle(error, net.IP, time.Time)
}

type Reply struct {
	Address net.IP
	Error   error
	Recieve time.Time
}

type simpleRequest struct {
	result error

	tStart  time.Time
	tFinish time.Time

	reply  chan Reply
	closed bool
	mtx    sync.RWMutex
}

func create() simpleRequest {
	return simpleRequest{
		tStart:  time.Now(),
		tFinish: time.Now(),
		reply:   make(chan Reply),
	}
}

func (req *simpleRequest) init() {
	req.reply = make(chan Reply)
	req.tStart = time.Now()
	req.tFinish = time.Now()
}

func (req *simpleRequest) close() {
	req.mtx.Lock()
	req.closed = true
	close(req.reply)
	req.mtx.Unlock()

}
func (req *simpleRequest) handler(err error, addr net.IP, tRecieve *time.Time) {

	req.mtx.RLock()
	req.result = err

	defer req.mtx.RUnlock()
	req.reply <- Reply{
		Address: addr,
		Error:   err,
		Recieve: *tRecieve,
	}
}

func (req simpleRequest) rrt() (time.Duration, error) {
	if req.result != nil {
		return 0, req.result
	}
	if req.tFinish == req.tStart {
		return 0, nil
	}
	return req.tFinish.Sub(req.tStart), nil

}

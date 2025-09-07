package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
)

type history struct {
	send     int
	received int
	lost     int
	results  []int64 // ring, start index = .received%len
	mtx      sync.RWMutex
}

func (h *history) getLost() string {
	h.mtx.RLock()
	defer h.mtx.RUnlock()
	return fmt.Sprintf("%d", (h.lost/h.send)*100)
}
func (h *history) getLast() string {
	h.mtx.RLock()
	defer h.mtx.RUnlock()
	length := len(h.results)
	if length == 0 {
		return "n/a"
	} else {
		return strconv.Itoa(int(h.results[length-1]))
	}
}
func (h *history) getBest() string {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	length := len(h.results)
	if length == 0 {
		return "n/a"
	} else {
		minVal := h.results[0]
		for i := 0; i < len(h.results); i++ {
			if h.results[i] < minVal {
				minVal = h.results[i] // Update if a smaller value is found
			}
		}
		return strconv.Itoa(int(minVal))
	}
}
func (h *history) getWorst() string {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	length := len(h.results)
	if length == 0 {
		return "n/a"
	} else {
		maxVal := h.results[0]
		for i := 0; i < len(h.results); i++ {
			if h.results[i] > maxVal {
				maxVal = h.results[i] // Update if a smaller value is found
			}
		}
		return strconv.Itoa(int(maxVal))
	}
}
func (h *history) getMean() string {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	length := len(h.results)
	if length == 0 {
		return "n/a"
	} else {
		sum := 0.0
		for _, num := range h.results {
			sum += float64(num)
		}
		mean := sum / float64(length)
		return strconv.Itoa(int(mean))
	}
}

type destination struct {
	host   string
	remote *net.IPAddr
}

type destination_cidr struct {
	ip       net.IP
	net      *net.IPNet
	type_net int
}

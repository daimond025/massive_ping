package main

import (
	"net"
	"sync"
	"time"
)

var (
	stats       map[string]history
	pinger_mass *Pinger
)

type Job struct {
	ID      int
	attemp  int
	timeout time.Duration
	Value   string
}

type WorkerPool struct {
	workersCount int
	jobsChannel  chan Job
	wg           sync.WaitGroup
	mutex        sync.Mutex
}

func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		workersCount: workers,
		jobsChannel:  make(chan Job, workers),
	}
}
func (p *WorkerPool) Start() {
	for i := 1; i <= p.workersCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *WorkerPool) AddJob(job Job) {
	p.jobsChannel <- job
}

func (p *WorkerPool) Stop() {
	close(p.jobsChannel)
	p.wg.Wait()
}

func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()
	for job := range p.jobsChannel {
		ip := net.ParseIP(job.Value)
		if ip == nil {
			continue
		}

		IPADRS := &net.IPAddr{IP: ip}
		rtt, err := pinger_mass.Ping(IPADRS, job.timeout, job.attemp)

		p.mutex.Lock()
		stats, ok_exist := pinger_mass.history[IPADRS.String()]
		if !ok_exist {
			stats = history{send: 0, received: 0, lost: 0, results: []int64{}}
		}

		if err == nil {
			//fmt.Printf("Process %s \n", job.Value)
			stats.send++
			stats.received++
			stats.results = append(stats.results, rtt.Milliseconds())
		} else {
			stats.send++
			stats.lost++
		}
		pinger_mass.history[IPADRS.String()] = stats
		p.mutex.Unlock()
	}
}

func setIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
}

func Ping_CIDR(pinger *Pinger, attempts int, poolSize int, timeout time.Duration) *Pinger {
	pinger_mass = pinger
	stats = make(map[string]history)

	pool := NewWorkerPool(poolSize)
	pool.Start()

	for _, dest := range pinger_mass.target_cidr {
		ip := dest.ip
		ipnet := dest.net

		for currentIP := ip.Mask(ipnet.Mask); ipnet.Contains(currentIP); setIP(currentIP) {
			for i := 0; i < attempts; i++ {
				pool.AddJob(Job{attemp: i, Value: currentIP.String(), timeout: timeout})
			}
		}
	}
	pool.Stop()
	return pinger_mass
}

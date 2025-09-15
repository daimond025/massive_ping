package ping

import (
	"time"
)

func (pinger *Pinger) add_rezult(dest destination, timeout time.Duration) {

}

// save result to list
func (pinger *Pinger) ping_item(dest destination, timeout time.Duration, attempt int) {
	rtt, err := pinger.Ping(dest.remote, timeout, attempt)

	pinger.stat_add.Lock()
	defer pinger.stat_add.Unlock()

	stats, ok_exist := pinger.history[dest.host]
	if !ok_exist {
		stats = history{send: 0, received: 0, lost: 0, results: []int64{}}
	}

	if err == nil {
		stats.send++
		stats.received++
		stats.results = append(stats.results, rtt.Milliseconds())
	} else {
		stats.send++
		stats.lost++
	}
	pinger.history[dest.host] = stats
	pinger.complate.Done()

}

func (pinger Pinger) PingRequest(timeout time.Duration, attempts int) {
	for _, dest := range pinger.target {
		for i := 0; i < attempts; i++ {
			pinger.complate.Add(1)
			go pinger.ping_item(dest, timeout, i)
		}
	}
	pinger.complate.Wait()
}

package ping

import "fmt"

func Out_std(p *Pinger) {
	size := uint(len(p.dataload))

	for adds, stats := range p.history {

		if stats.received == 0 {
			continue
		}
		out := fmt.Sprintf("Host: %s; recieve bytes %v,  send - %v, receive - %v; statistics min -  %s, max - %s, mean  - %s",
			adds, size, stats.send, stats.received, stats.getBest(), stats.getWorst(), stats.getMean())
		fmt.Println(out)
	}
}

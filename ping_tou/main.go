package main

import (
	"flag"
	"fmt"
	ping "github.com/daimond025/massive_ping"
	"os"
	"time"
)

var (
	attempts    uint = 3
	timeout          = time.Second
	bind_v6          = "::"
	bind_v4          = "0.0.0.0"
	destination      = ""
	size        uint = 56
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "[options] host [host [...]]")
		flag.PrintDefaults()
	}

	flag.UintVar(&attempts, "a", attempts, "number of attempts")
	flag.DurationVar(&timeout, "t", timeout, "timeout for a echo request")
	flag.UintVar(&size, "s", size, "size of additional payload data")

	flag.StringVar(&destination, "h", destination, "size of additional payload data")
	flag.StringVar(&bind_v4, "4", bind_v4, "IPv4 bind address")
	flag.StringVar(&bind_v6, "6", bind_v6, "IPv6 bind address")
	flag.Parse()

	//destination := " 8.8.8.8, google.com, fe80:0000:0000:0000:0f19:1faf:008:5010"
	p, err := ping.NewPinger()
	if err != nil {
		panic(err)
	}
	targets := p.Targets(destination)
	if targets

	err = p.CreateConnection(bind_v4, bind_v6, size)
	if err != nil {
		panic(err)
	}

	go p.PingRequest(timeout, int(attempts))

	tui := ping.BuildTUI(p)
	go tui.Update(p, 1*time.Second)
	if err := tui.Run(); err != nil {
		panic(err)
	}
	defer p.Close()
}

package main

import (
	"flag"
	"fmt"
	ping "github.com/daimond025/massive_ping"
	"os"
	"time"
)

var (
	attempts       uint = 3
	timeout             = time.Second
	bind_v6             = "::"
	bind_v4             = "0.0.0.0"
	proto4, proto6 bool
	size           uint = 56
	bind           string
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "[options] host [host [...]]")
		flag.PrintDefaults()
	}

	flag.UintVar(&attempts, "attempts", attempts, "number of attempts")
	flag.DurationVar(&timeout, "timeout", timeout, "timeout for a echo request")
	flag.UintVar(&size, "s", size, "size of additional payload data")
	flag.StringVar(&bind_v4, "4", bind_v4, "IPv4 bind address")
	flag.StringVar(&bind_v6, "6", bind_v6, "IPv6 bind address")
	flag.BoolVar(&proto4, "4", proto4, "use IPv4 (mutually exclusive with -6)")
	flag.BoolVar(&proto6, "6", proto6, "use IPv6 (mutually exclusive with -4)")
	flag.StringVar(&bind, "bind", "", "IPv4 or IPv6 bind address (defaults to 0.0.0.0 for IPv4 and :: for IPv6)")
	flag.Parse()

	bind_v4 := "0.0.0.0"
	bind_v6 := "::"

	//destination := " 127.0.0.1 fe80:0000:0000:0000:0f19:1faf:008:5010 8.8.8.8 google.com"
	destination := " 8.8.8.8 google.com fe80:0000:0000:0000:0f19:1faf:008:5010"
	p, err := ping.NewPinger()
	if err != nil {
		panic(err)
	}
	p.Targets(destination)

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

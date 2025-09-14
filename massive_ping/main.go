package main

import (
	"flag"
	"fmt"
	ping "github.com/daimond025/massive_ping"
	"github.com/digineo/go-logwrap"
	"os"
	"runtime"
	"time"
)

var (
	log = &logwrap.Instance{}

	timeout   = 5 * time.Second
	attempts  = 1
	poolSize  = 3 * runtime.NumCPU()
	bind_v6   = "::"
	bind_v4   = "0.0.0.0"
	size      = uint(56)
	verbose   bool
	cidr_adds = ""
	pinger    *ping.Pinger
)

func main() {

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "[options] CIDR [CIDR [...]]")
		flag.PrintDefaults()
	}
	flag.IntVar(&attempts, "c", attempts, "number of ping attempts per address")
	flag.DurationVar(&timeout, "w", timeout, "timeout for a single echo request")
	flag.UintVar(&size, "s", size, "size of additional payload data")
	flag.StringVar(&bind_v4, "4", bind_v4, "IPv4 bind address")
	flag.StringVar(&bind_v6, "6", bind_v6, "IPv6 bind address")
	flag.IntVar(&poolSize, "P", poolSize, "concurrency level")
	flag.BoolVar(&verbose, "v", verbose, "also print out unreachable addresses")
	flag.StringVar(&cidr_adds, "cidr", cidr_adds, "set cidr network  ")
	flag.Parse()

	if bind_v4 == "" && bind_v6 == "" {
		log.Errorf("need at least an IPv4 (-bind4 flag) or IPv6 (-bind6 flag) address to bind to")
		os.Exit(0)
	}

	if attempts <= 0 {
		log.Errorf("number of ping attempts (-c flag) must be > 0")
		os.Exit(0)
	}

	if cidr_adds == "" {
		log.Errorf("set CIDR networks  (-cidr flag) must be not empty")
		os.Exit(0)
	}

	p, err := ping.NewPinger()

	if err != nil {
		panic(err)
	}
	err = p.CreateConnection(bind_v4, bind_v6, size)
	if err != nil {
		panic(err)
	}

	//cidr_adds := " 192.138.88.1/24 192.138.89.1/24 2001:db8::/32"
	//cidr_adds := " 192.138.89.1/24 192.168.1.1/28 "
	err_cidr := p.Targets_CIDR(cidr_adds)
	if err_cidr != nil {
		panic(err_cidr)
	}
	p = ping.Ping_CIDR(p, attempts, poolSize, timeout)

	ping.Out_std(p)

}

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

	timeout  = 5 * time.Second
	attempts = 1
	poolSize = 3 * runtime.NumCPU()
	interval = 100 * time.Millisecond
	ifname   = ""
	bind_v6  = "::"
	bind_v4  = "0.0.0.0"
	size     = uint(56)
	force    bool
	verbose  bool
	pinger   *ping.Pinger
)

func main() {

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "[options] CIDR [CIDR [...]]")
		flag.PrintDefaults()
	}
	flag.IntVar(&attempts, "c", attempts, "number of ping attempts per address")
	flag.DurationVar(&timeout, "w", timeout, "timeout for a single echo request")
	flag.DurationVar(&interval, "i", interval, "CIDR iteration interval")
	flag.UintVar(&size, "s", size, "size of additional payload data")
	flag.StringVar(&bind_v4, "4", bind_v4, "IPv4 bind address")
	flag.StringVar(&bind_v6, "6", bind_v6, "IPv6 bind address")
	flag.StringVar(&ifname, "I", ifname, "interface name/IPv6 zone")
	flag.IntVar(&poolSize, "P", poolSize, "concurrency level")
	flag.BoolVar(&force, "f", force, "sanity flag needed if you want to ping more than 4096 hosts (/20)")
	flag.BoolVar(&verbose, "v", verbose, "also print out unreachable addresses")
	flag.Parse()

	if bind_v4 == "" && bind_v6 == "" {
		log.Errorf("need at least an IPv4 (-bind4 flag) or IPv6 (-bind6 flag) address to bind to")
	}

	if attempts <= 0 {
		log.Errorf("number of ping attempts (-c flag) must be > 0")
	}

	p, err := ping.NewPinger()

	if err != nil {
		panic(err)
	}
	err = p.CreateConnection(bind_v4, bind_v6)
	if err != nil {
		panic(err)
	}

	//destination := " 192.168.1.7 "
	//p.Targets(destination)
	//p.PingRequest(timeout, int(attempts))

	//destination := " 192.138.88.1/24 192.138.89.1/24 2001:db8::/32"
	destination := " 192.168.1.1/28 "
	p.Targets_CIDR(destination)
	p = ping.Ping_CIDR(p, attempts, poolSize, timeout)

	ping.Out_std(p)

	//ip, ipnet, err := net.ParseCIDR("192.168.1.0/28")
	//
	//jobs := make(chan net.IP, 50)
	//var wg sync.WaitGroup
	//for i := 1; i <= poolSize; i++ {
	//	wg.Add(1)
	//	go worker(i, jobs, &wg)
	//}
	//
	//for currentIP := ip.Mask(ipnet.Mask); ipnet.Contains(currentIP); nextIP(currentIP) {
	//	fmt.Println(currentIP)
	//	jobs <- currentIP
	//}
	//
	//close(jobs)
	//wg.Wait()
	//fmt.Println("Все IP-адреса обработаны.")

	//for _, cidr := range flag.Args() {
	//	ip, ipnet, err := net.ParseCIDR(cidr)
	//	if err != nil {
	//		log.Println(err)
	//		continue
	//	}
	//	w := &workGenerator{ip: ip, net: ipnet}
	//	generator = append(generator, w)
	//	total += w.size()
	//}
}

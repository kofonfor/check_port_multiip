package main

import (
	"flag"
	"fmt"
	dns "github.com/Focinfi/go-dns-resolver"
	//"log"
	"net"
	"os"
	"sync"
	"time"
)

type CommandLineConfig struct {
	host_name *string
	port      *int64
}

func (*CommandLineConfig) Parse() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
}

var commandLineCfg = CommandLineConfig{
	host_name: flag.String("host_name", "localhost", "An FQDN to check"),
	port:      flag.Int64("port", 22, "A port to check"),
}

func main() {
	commandLineCfg.Parse()
	dns.Config.SetTimeout(uint(2))
	dns.Config.RetryTimes = uint(4)
	all_ips := make([]string, 0)
	dead_ips := make([]string, 0)
	var wg sync.WaitGroup

	if results, err := dns.Exchange(*commandLineCfg.host_name, "8.8.8.8:53", dns.TypeA); err == nil {
		for _, r := range results {
			all_ips = append(all_ips, r.Content)
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%v", r.Content, *commandLineCfg.port), 2*time.Second)
				if err != nil {
					dead_ips = append(dead_ips, r.Content)
				}
			}()
		}
	} else {
		fmt.Printf("CRITICAL - %v\n", err)
		os.Exit(3)
	}
	wg.Wait()
	if len(dead_ips) != 0 {
		fmt.Printf("CRITICAL - some probes (%v/%v) failed\n", len(dead_ips), len(all_ips))
		os.Exit(3)
	} else {
		fmt.Printf("OK - all probes succeeded\n")
		os.Exit(0)
	}
}

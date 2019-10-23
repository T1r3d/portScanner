package main

import (
	"fmt"
	"net"
	"sync"
	"github.com/sparrc/go-ping"
	"time"
	"strings"
	"strconv"
	"flag"
)

func pingDetect(ip string) bool {
	stats := make(chan *ping.Statistics, 1)
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		panic(err)
	}
	pinger.Count = 4
	pinger.SetPrivileged(true)
	go func() {
		fmt.Println("Start ping test...")
		pinger.Run()	
		stats <- pinger.Statistics()
	}()
	select {
	case state := <-stats:
		fmt.Println(state)
		return true
	case <-time.After(5 * time.Second):
		fmt.Println("Ping Timeout: The target is not alive.")
		return false
	}
}

func tcpConnectDetect(ip, port string, wg *sync.WaitGroup) bool {
	target := strings.Join([]string{ip, port}, ":")
	//fmt.Printf("Scanning %s\n", target)
	_, err := net.Dial("tcp", target)
	if err != nil {
		wg.Done()
		return false
	}
	fmt.Printf("[+] Port %s open\n", port)
	wg.Done()
	return true

}

func main() {
	//command-line args parse
	iptr := flag.String("ip", "127.0.0.1", "Target IP")	
	portsptr := flag.String("ports", "80,443", "Target ports")
	flag.Parse()	

	//detect if the host alive
	pingDetect(*iptr)

	var wg sync.WaitGroup

	ports := strings.Split(*portsptr, ",")
	
	for _, ran := range ports {
		port := strings.Split(ran, "-")
		if len(port) == 1 {
			wg.Add(1)
			go tcpConnectDetect(*iptr, port[0], &wg)
		} else {
			start, _ := strconv.Atoi(port[0])
			end, _ := strconv.Atoi(port[1])
			for p := start; p <= end; p++ {
				wg.Add(1)
				go tcpConnectDetect(*iptr, strconv.Itoa(p), &wg)
			}
		}
	}

	wg.Wait()
	fmt.Println("[*] Scan Finished")
	
}

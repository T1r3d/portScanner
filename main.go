package main

import (
	"fmt"
	"net"
	"sync"
	"github.com/sparrc/go-ping"
	"time"
	"strings"
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
	for _, port := range ports {
		wg.Add(1)
		go tcpConnectDetect(*iptr, port, &wg)
	}

	wg.Wait()
	/*
	conn, err := net.Dial("tcp", "120.79.219.67:8000")
	if err != nil {
		fmt.Println("Error!")
		fmt.Println(err)
		fmt.Println("Error2!")
		//panic(err)
		os.Exit(2)
	}
	fmt.Fprintf(conn, "GET / HTTP/1.1\r\n\r\n")
	status, err := bufio.NewReader(conn).ReadString('\n')
	fmt.Println(status)
	*/
}

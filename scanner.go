package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sparrc/go-ping"
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
		fmt.Println("[*] Start ping test...")
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

func tcpConnectDetect(ip, port string) bool {
	target := strings.Join([]string{ip, port}, ":")
	//fmt.Printf("Scanning %s\n", target)
	_, err := net.Dial("tcp", target)
	if err != nil {
		return false
	}
	fmt.Printf("[+] Port %s open\n", port)
	return true
}

func scanner(id int, ip string, ports <-chan string, wg *sync.WaitGroup) {
	for port := range ports {
		fmt.Printf("[*] scanner-%d receive port %s\n", id, port)
		tcpConnectDetect(ip, port)			
		wg.Done()
	}
}

func banner() {
	fmt.Println(`                                 /$$      /$$$$$$                                                             
                                | $$     /$$__  $$                                                            
  /$$$$$$   /$$$$$$   /$$$$$$  /$$$$$$  | $$  \__/  /$$$$$$$  /$$$$$$  /$$$$$$$  /$$$$$$$   /$$$$$$   /$$$$$$ 
 /$$__  $$ /$$__  $$ /$$__  $$|_  $$_/  |  $$$$$$  /$$_____/ |____  $$| $$__  $$| $$__  $$ /$$__  $$ /$$__  $$
| $$  \ $$| $$  \ $$| $$  \__/  | $$     \____  $$| $$        /$$$$$$$| $$  \ $$| $$  \ $$| $$$$$$$$| $$  \__/
| $$  | $$| $$  | $$| $$        | $$ /$$ /$$  \ $$| $$       /$$__  $$| $$  | $$| $$  | $$| $$_____/| $$      
| $$$$$$$/|  $$$$$$/| $$        |  $$$$/|  $$$$$$/|  $$$$$$$|  $$$$$$$| $$  | $$| $$  | $$|  $$$$$$$| $$      
| $$____/  \______/ |__/         \___/   \______/  \_______/ \_______/|__/  |__/|__/  |__/ \_______/|__/      
| $$                                                                                                          
| $$                                                                                                          
|__/                                                                                               Author: t1r3d`)
}

func main() {
	banner()
	//Command-line args parse.
	iptr := flag.String("ip", "127.0.0.1", "Target IP")
	portsptr := flag.String("ports", "80,443", "Target ports")
	flag.Parse()
	//Detect if the host alive.
	pingDetect(*iptr)
	//Start concurrency tcp connect detect.
	/*
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
	*/
	//Start concurrency tcp connect detect by workerpool.
	var wg sync.WaitGroup
	ports := make(chan string, 1000)
	for i := 0; i < 1000; i++ {
		go scanner(i, *iptr, ports, &wg)
	}	
	portsList := strings.Split(*portsptr, ",")
	for _, ran := range portsList {
		port := strings.Split(ran, "-")
		if len(port) == 1 {
			wg.Add(1)
			ports <- port[0]
		} else {
			start, _ := strconv.Atoi(port[0])
			end, _ := strconv.Atoi(port[1])
			for p := start; p <= end; p++ {
				wg.Add(1)
				ports <- strconv.Itoa(p)
			}
		}
	}

	wg.Wait()
	fmt.Println("[*] Scan Finished")
}

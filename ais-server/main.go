package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

var (
	maxPositionSleep = flag.Float64("s", 3.0, "The max number of seconds to sleep between positions")
	positionChan     = make(chan string)
)

func main() {
	flag.Parse()
	rand.Seed(time.Now().UTC().UnixNano())

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	positionScanner := bufio.NewScanner(file)
	go func() {
		for positionScanner.Scan() {
			positionChan <- positionScanner.Text()
			sleepyTime := rand.Float64() * *maxPositionSleep * float64(time.Second.Nanoseconds())
			time.Sleep(time.Duration(sleepyTime))
		}
	}()

	fmt.Println("Launching server...")
	ln, _ := net.Listen("tcp4", "127.0.0.1:10110")
	conn, _ := ln.Accept()
	defer conn.Close()

	for {
		select {
		case position := <-positionChan:
			fmt.Println(position)
			_, _ = conn.Write([]byte(position))
			_, _ = conn.Write([]byte("\n"))
		}
	}
}

package main

import (
	"flag"
	"os"
	"log"
	"bufio"
	"time"
	"math/rand"
	"fmt"
	"net"
)

var (
	positionFile     = flag.String("f", "", "The name of the file to read")
	maxPositionSleep = flag.Int("s", 10, "The max number of seconds to sleep between positions")
	positionChan     = make(chan string)
)

func main() {
	flag.Parse()
	rand.Seed(time.Now().UTC().UnixNano())

	file, err := os.Open(*positionFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	positionScanner := bufio.NewScanner(file)
	go func() {
		for positionScanner.Scan() {
			positionChan <- positionScanner.Text()
			time.Sleep(time.Duration(rand.Intn(*maxPositionSleep) * int(time.Second.Nanoseconds())))
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
			conn.Write([]byte(position))
			conn.Write([]byte("\n"))
		}
	}
}

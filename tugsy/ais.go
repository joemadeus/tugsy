package main

import (
	"bufio"
	"net"
	"time"
)

type RemoteAISServer struct {
	// The "in" channel on aislib.Router()
	RouterIn chan string
	hostname string
	port     int
	user     string
	pass     string
}

func NewRemoteAISServer(host string, port int, user string, pass string, in chan string) (*RemoteAISServer, error) {
	server := &RemoteAISServer{
		RouterIn: in,
		hostname: host,
		port:     port,
		user:     user,
		pass:     pass,
	}

	go func() {
		sleep := 10 // How many seconds to sleep after a timeout
		timeoutSleep := time.Duration(sleep) * time.Second
		for {
			serverAddr, err := net.ResolveTCPAddr("tcp", server.hostname)
			if err != nil {
				logger.Error("Could not resolve the AIS host", "error", err, "host", server.hostname, "retrying in", sleep)
				time.Sleep(timeoutSleep)
				continue
			}

			conn, err := net.DialTCP("tcp", nil, serverAddr)
			if err != nil {
				logger.Error("Could not connect to the AIS host", "error", err, "host", server.hostname, "retrying in", sleep)
				time.Sleep(timeoutSleep)
				continue
			}
			defer conn.Close()

			connbuf := bufio.NewScanner(conn)
			connbuf.Split(bufio.ScanLines)
			for connbuf.Scan() {
				in <- connbuf.Text()
				conn.SetReadDeadline(time.Now().Add(15 * time.Second))
			}

			logger.Error("Connection broken", "host", server.hostname, "retrying in", sleep)
			time.Sleep(timeoutSleep)
		}
	}()

	return server, nil
}

package main

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"errors"

	"github.com/andmarios/aislib"
)

const (
	connRetryTimeoutSecs = 10 // How many seconds to sleep after a timeout
	readDeadlineSecs     = 15
)

var (
	NoRouterConfigFound = errors.New("Could not find router configs")
)

type SourcedMessage struct {
	*aislib.Message
	sourceName string
}

type RemoteAISServer struct {
	inStrings chan string
	Decoded   chan SourcedMessage
	Failed    chan aislib.FailedSentence
	conn      net.Conn

	sourceName string

	hostname string
	port     int
}

func RemoteAISServersFromConfig(decoded chan SourcedMessage, failed chan aislib.FailedSentence, config *Config) ([]*RemoteAISServer, error) {
	if config.IsSet("routers") == false {
		return nil, NoRouterConfigFound
	}

	routerConfigs := config.GetStringMap("routers")
	routers := make(map[string]*RemoteAISServer)
	for sourceName, routerConfig := range routerConfigs {
		routers[sourceName] = &RemoteAISServer{
			inStrings: make(chan string),
			Decoded:   decoded,
			Failed:    failed,

			sourceName: sourceName,

			hostname: host,
			port:     port,
		}
	}

	return nil, nil
}

func (router *RemoteAISServer) start() {
	decoded := make(chan aislib.Message)
	go aislib.Router(router.inStrings, decoded, router.Failed)

	// "decorate" incoming, generic messages with this router's special sauce
	go func() {
		var m aislib.Message
		for {
			m = <-decoded
			router.Decoded <- SourcedMessage{&m, router.sourceName}
		}
	}()

	go func() {
		timeoutSleep := time.Duration(connRetryTimeoutSecs) * time.Second
		for running {
			if router.conn != nil {
				logger.Error("Connection broken", "host", router.hostname, "retrying in", connRetryTimeoutSecs)
				time.Sleep(timeoutSleep)
			}

			serverAddr, err := net.ResolveTCPAddr("tcp", router.hostname)
			if err != nil {
				logger.Error("Could not resolve the AIS host", "error", err, "host", router.hostname, "retrying in", connRetryTimeoutSecs)
				continue
			}

			router.conn, err = net.DialTCP("tcp", nil, serverAddr)
			if err != nil {
				logger.Error("Could not connect to the AIS host", "error", err, "host", router.hostname, "retrying in", connRetryTimeoutSecs)
				continue
			}
			defer router.conn.Close()

			connbuf := bufio.NewScanner(router.conn)
			connbuf.Split(bufio.ScanLines)
			for connbuf.Scan() && running {
				router.inStrings <- connbuf.Text()
				router.conn.SetReadDeadline(time.Now().Add(readDeadlineSecs * time.Second))
			}
		}
		logger.Info("Router reconnect loop exiting", "running", running)
	}()
}

func (router *RemoteAISServer) stop() error {
	return nil
}

func masterBlaster(decoded chan SourcedMessage, failed chan aislib.FailedSentence) {
	for {
		select {
		case message := <-decoded:
			switch message.Type {
			case 1, 2, 3:
				t, err := aislib.DecodeClassAPositionReport(message.Payload)
				if err != nil {
					logger.Error("Decoding class A report", "err", err)
				}
				fmt.Println(t)
			case 4:
				t, err := aislib.DecodeBaseStationReport(message.Payload)
				if err != nil {
					logger.Error("Decoding base station report", "err", err)
				}
				fmt.Println(t)
			case 5:
				t, err := aislib.DecodeStaticVoyageData(message.Payload)
				if err != nil {
					logger.Error("Decoding voyage data", "err", err)
				}
				fmt.Println(t)
			case 8:
				t, err := aislib.DecodeBinaryBroadcast(message.Payload)
				if err != nil {
					logger.Error("Decoding binary broadcast", "err", err)
				}
				fmt.Println(t)
			case 18:
				t, err := aislib.DecodeClassBPositionReport(message.Payload)
				if err != nil {
					logger.Error("Decoding class B report", "err", err)
				}
				fmt.Println(t)
			default:
				logger.Debug("Unsupported message type %2d", message.Type)
			}

		case problematic := <-failed:
			logger.Debug("Failed message", "issue", problematic.Issue, "sentence", problematic.Sentence)
		}
	}
}

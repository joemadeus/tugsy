package main

import (
	"bufio"
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

type SourcedClassAPositionReport struct {
	*aislib.ClassAPositionReport
	sourceName string
}

type SourcedClassBPositionReport struct {
	*aislib.ClassBPositionReport
	sourceName string
}

type SourcedBaseStationReport struct {
	*aislib.BaseStationReport
	sourceName string
}

type SourcedBinaryBroadcast struct {
	*aislib.BinaryBroadcast
	sourceName string
}

type SourcedStaticVoyageData struct {
	*aislib.StaticVoyageData
	sourceName string
}

type RemoteAISServer struct {
	inStrings chan string
	Decoded   chan aislib.Message
	Failed    chan aislib.FailedSentence
	conn      net.Conn

	sourceName string

	hostname string
	port     int
}

func RemoteAISServersFromConfig(decoded chan aislib.Message, failed chan aislib.FailedSentence, config *Config) ([]*RemoteAISServer, error) {
	if config.IsSet("routers") == false {
		return nil, NoRouterConfigFound
	}

	var routers []*RemoteAISServer
	routerConfigs := config.Get("routers")
	for routerConfig, i := range routerConfigs {
		routers[i] = &RemoteAISServer{
			inStrings: make(chan string),
			Decoded:   decoded,
			Failed:    failed,

			sourceName: sourceName,

			hostname: host,
			port:     port,
		}
	}

	return routers, nil
}

func (router *RemoteAISServer) start() {
	decoded := make(chan aislib.Message)
	go aislib.Router(router.inStrings, decoded, router.Failed)

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

func (router *RemoteAISServer) DecodePositions(decoded chan aislib.Message, failed chan aislib.FailedSentence) {
	logger.Info("Starting AIS loop for " + router.sourceName)
	for {
		select {
		case message := <-decoded:
			switch message.Type {
			case 1, 2, 3:
				t, err := aislib.DecodeClassAPositionReport(message.Payload)
				if err != nil {
					logger.Error("Decoding class A report", "err", err)
					break
				}
				report := &SourcedClassAPositionReport{&t, router.sourceName}
				TheData.AddAPosition(report)

			case 4:
				t, err := aislib.DecodeBaseStationReport(message.Payload)
				if err != nil {
					logger.Error("Decoding base station report", "err", err)
					break
				}
				report := &SourcedBaseStationReport{&t, router.sourceName}
				TheData.UpdateBaseStationReport(report)

			case 5:
				t, err := aislib.DecodeStaticVoyageData(message.Payload)
				if err != nil {
					logger.Error("Decoding voyage data", "err", err)
					break
				}
				report := &SourcedStaticVoyageData{&t, router.sourceName}
				TheData.UpdateStaticVoyageData(report)

			case 8:
				t, err := aislib.DecodeBinaryBroadcast(message.Payload)
				if err != nil {
					logger.Error("Decoding binary broadcast", "err", err)
					break
				}
				report := &SourcedBinaryBroadcast{&t, router.sourceName}
				TheData.UpdateBinaryBroadcast(report)

			case 18:
				t, err := aislib.DecodeClassBPositionReport(message.Payload)
				if err != nil {
					logger.Error("Decoding class B report", "err", err)
					break
				}
				report := &SourcedClassBPositionReport{&t, router.sourceName}
				TheData.AddBPosition(report)

			default:
				logger.Debug("Unsupported message type %2d", message.Type)
			}

		case problematic := <-failed:
			logger.Debug("Failed message", "issue", problematic.Issue, "sentence", problematic.Sentence)
		}
	}
}

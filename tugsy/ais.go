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
	NoRouterConfigFound = errors.New("could not find router configs")
)

type Positionable interface {
	GetPositionReport() *aislib.PositionReport
	GetSource() string
	GetReceivedTime() time.Time
}

type SourceAndTime struct {
	sourceName   string
	receivedTime time.Time
}

func (st *SourceAndTime) GetSource() string {
	return st.sourceName
}

func (st *SourceAndTime) GetReceivedTime() time.Time {
	return st.receivedTime
}

type SourcedClassAPositionReport struct {
	aislib.ClassAPositionReport
	SourceAndTime
}

func (aPos *SourcedClassAPositionReport) GetPositionReport() *aislib.PositionReport {
	return &aPos.PositionReport
}

type SourcedClassBPositionReport struct {
	aislib.ClassBPositionReport
	SourceAndTime
}

func (bPos *SourcedClassBPositionReport) GetPositionReport() *aislib.PositionReport {
	return &bPos.PositionReport
}

type SourcedBaseStationReport struct {
	aislib.BaseStationReport
	SourceAndTime
}

type SourcedBinaryBroadcast struct {
	aislib.BinaryBroadcast
	SourceAndTime
}

type SourcedStaticVoyageData struct {
	aislib.StaticVoyageData
	SourceAndTime
}

type RemoteAISServer struct {
	inStrings chan string
	Decoded   chan aislib.Message
	Failed    chan aislib.FailedSentence
	conn      net.Conn

	SourceName string
	Host       string
	Port       int
}

func RemoteAISServersFromConfig(decoded chan aislib.Message, failed chan aislib.FailedSentence, config *Config) ([]*RemoteAISServer, error) {
	if config.IsSet("routers") == false {
		return nil, NoRouterConfigFound
	}

	var routers []*RemoteAISServer
	err := config.UnmarshalKey("routers", &routers)
	if err != nil {
		return nil, err
	}

	for _, router := range routers {
		router.Decoded = decoded
		router.Failed = failed
		router.inStrings = make(chan string)
	}

	return routers, nil
}

func (router *RemoteAISServer) start() {
	decoded := make(chan aislib.Message)
	go aislib.Router(router.inStrings, decoded, router.Failed)

	go func() {
		timeoutSleep := time.Duration(connRetryTimeoutSecs) * time.Second
		for MachineAndProcessState.running {
			if router.conn != nil {
				logger.Warn("Connection broken", "host", router.Host, "retrying in", connRetryTimeoutSecs)
				time.Sleep(timeoutSleep)
			}

			serverAddr, err := net.ResolveTCPAddr("tcp", router.Host)
			if err != nil {
				logger.Error("Could not resolve the AIS host", "error", err, "host", router.Host, "retrying in", connRetryTimeoutSecs)
				continue
			}

			router.conn, err = net.DialTCP("tcp", nil, serverAddr)
			if err != nil {
				logger.Error("Could not connect to the AIS host", "error", err, "host", router.Host, "retrying in", connRetryTimeoutSecs)
				continue
			}
			defer router.conn.Close()

			connbuf := bufio.NewScanner(router.conn)
			connbuf.Split(bufio.ScanLines)
			for connbuf.Scan() && MachineAndProcessState.running {
				router.inStrings <- connbuf.Text()
				router.conn.SetReadDeadline(time.Now().Add(readDeadlineSecs * time.Second))
			}
		}
		logger.Info("Router reconnect loop exiting", "running", MachineAndProcessState.running)
	}()
}

func (router *RemoteAISServer) stop() error {
	return nil
}

func (router *RemoteAISServer) DecodePositions(decoded chan aislib.Message, failed chan aislib.FailedSentence) {
	logger.Info("Starting AIS loop for " + router.SourceName)
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
				report := &SourcedClassAPositionReport{t, SourceAndTime{router.SourceName, time.Now()}}
				MachineAndProcessState.TheData.AddPosition(report)

			case 4:
				t, err := aislib.DecodeBaseStationReport(message.Payload)
				if err != nil {
					logger.Error("Decoding base station report", "err", err)
					break
				}
				report := &SourcedBaseStationReport{t, SourceAndTime{router.SourceName, time.Now()}}
				MachineAndProcessState.TheData.UpdateBaseStationReport(report)

			case 5:
				t, err := aislib.DecodeStaticVoyageData(message.Payload)
				if err != nil {
					logger.Error("Decoding voyage data", "err", err)
					break
				}
				report := &SourcedStaticVoyageData{t, SourceAndTime{router.SourceName, time.Now()}}
				MachineAndProcessState.TheData.UpdateStaticVoyageData(report)

			case 8:
				t, err := aislib.DecodeBinaryBroadcast(message.Payload)
				if err != nil {
					logger.Error("Decoding binary broadcast", "err", err)
					break
				}
				report := &SourcedBinaryBroadcast{t, SourceAndTime{router.SourceName, time.Now()}}
				MachineAndProcessState.TheData.UpdateBinaryBroadcast(report)

			case 18:
				t, err := aislib.DecodeClassBPositionReport(message.Payload)
				if err != nil {
					logger.Error("Decoding class B report", "err", err)
					break
				}
				report := &SourcedClassBPositionReport{t, SourceAndTime{router.SourceName, time.Now()}}
				MachineAndProcessState.TheData.AddPosition(report)

			default:
				logger.Debug("Unsupported message type %2d", message.Type)
			}

		case problematic := <-failed:
			logger.Debug("Failed message", "issue", problematic.Issue, "sentence", problematic.Sentence)
		}
	}
}

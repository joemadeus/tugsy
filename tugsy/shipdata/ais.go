package shipdata

import (
	"bufio"
	"net"
	"time"

	"errors"

	"github.com/andmarios/aislib"
	"github.com/joemadeus/tugsy/tugsy/config"
)

const (
	connRetryTimeoutSecs = 10 // How many seconds to sleep after a connection failure
	connRetryAttempts    = 10 // How many times to try a connection before failing it
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

	SourceName    string
	HostColonPort string
	conn          net.Conn
	connAttempts  uint
}

func RemoteAISServersFromConfig(decoded chan aislib.Message, failed chan aislib.FailedSentence, config *config.Config) ([]*RemoteAISServer, error) {
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

func (router *RemoteAISServer) Start() {
	go aislib.Router(router.inStrings, router.Decoded, router.Failed)

	timeoutSleep := time.Duration(connRetryTimeoutSecs) * time.Second
	go func() {
		for MachineAndProcessState.running {
			serverAddr, err := net.ResolveTCPAddr("tcp", router.HostColonPort)
			if err != nil {
				logger.Warn("Could not resolve an AIS host", "error", err, "host", router.HostColonPort, "retrying in", connRetryTimeoutSecs)
				router.connAttempts++
				if router.connAttempts > connRetryAttempts {
					logger.Error("Failing this AIS server", "host", router.HostColonPort)
					return
				}
				time.Sleep(timeoutSleep)
				continue
			}
			logger.Info("Resolved", "host", router.HostColonPort)

			router.conn, err = net.DialTCP("tcp", nil, serverAddr)
			if err != nil {
				logger.Warn("Could not connect to an AIS host", "error", err, "host", router.HostColonPort, "retrying in", connRetryTimeoutSecs)
				router.connAttempts++
				if router.connAttempts > connRetryAttempts {
					logger.Error("Failing this AIS server", "host", router.HostColonPort)
					return
				}
				time.Sleep(timeoutSleep)
				continue
			}
			logger.Info("Dialed", "host", router.HostColonPort)

			router.connAttempts = 0

			connbuf := bufio.NewScanner(router.conn)
			connbuf.Split(bufio.ScanLines)
			for connbuf.Scan() && MachineAndProcessState.running {
				router.inStrings <- connbuf.Text()
			}

			router.conn.Close()
			logger.Warn("Connection broken/not established", "host", router.HostColonPort, "retrying in", connRetryTimeoutSecs)
			time.Sleep(timeoutSleep)
		}
		logger.Info("Router reconnect loop exiting")
	}()
}

func (router *RemoteAISServer) DecodePositions(decoded chan aislib.Message, failed chan aislib.FailedSentence) {
	logger.Info("Starting AIS loop ", "source", router.SourceName)
	for {
		select {
		case message := <-decoded:
			switch message.Type {
			case 1, 2, 3:
				t, err := aislib.DecodeClassAPositionReport(message.Payload)
				if err != nil {
					logger.Warn("Decoding class A report", "err", err)
					break
				}
				report := &SourcedClassAPositionReport{t, SourceAndTime{router.SourceName, time.Now()}}
				logger.Trace("New type A position", "position", report)
				PositionData.AddPosition(report)

			case 4:
				t, err := aislib.DecodeBaseStationReport(message.Payload)
				if err != nil {
					logger.Warn("Decoding base station report", "err", err)
					break
				}
				report := &SourcedBaseStationReport{t, SourceAndTime{router.SourceName, time.Now()}}
				logger.Trace("New base station data", "data", report)
				PositionData.UpdateBaseStationReport(report)

			case 5:
				t, err := aislib.DecodeStaticVoyageData(message.Payload)
				if err != nil {
					logger.Warn("Decoding voyage data", "err", err)
					break
				}
				report := &SourcedStaticVoyageData{t, SourceAndTime{router.SourceName, time.Now()}}
				logger.Trace("New voyage data", "data", report)
				PositionData.UpdateStaticVoyageData(report)

			case 8:
				t, err := aislib.DecodeBinaryBroadcast(message.Payload)
				if err != nil {
					logger.Warn("Decoding binary broadcast", "err", err)
					break
				}
				report := &SourcedBinaryBroadcast{t, SourceAndTime{router.SourceName, time.Now()}}
				logger.Trace("New binary broadcast", "data", report)
				PositionData.UpdateBinaryBroadcast(report)

			case 18:
				t, err := aislib.DecodeClassBPositionReport(message.Payload)
				if err != nil {
					logger.Warn("Decoding class B report", "err", err)
					break
				}
				report := &SourcedClassBPositionReport{t, SourceAndTime{router.SourceName, time.Now()}}
				logger.Trace("New type B position", "position", report)
				PositionData.AddPosition(report)

			default:
				logger.Debug("Unsupported message type", "type", message.Type)
			}

		case problematic := <-failed:
			logger.Debug("Failed message", "issue", problematic.Issue, "sentence", problematic.Sentence)
		}
	}
}

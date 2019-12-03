package shipdata

import (
	"bufio"
	"errors"
	"net"
	"time"

	"github.com/andmarios/aislib"
	"github.com/joemadeus/tugsy/tugsy/config"
	logger "github.com/sirupsen/logrus"
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

	aisData      *AISData
	conn         net.Conn
	connAttempts uint
	running      bool
}

func RemoteAISServersFromConfig(aisdata *AISData, decoded chan aislib.Message, failed chan aislib.FailedSentence, config *config.Config) ([]*RemoteAISServer, error) {
	if config.IsSet("routers") == false {
		return nil, NoRouterConfigFound
	}

	var routers []*RemoteAISServer
	err := config.UnmarshalKey("routers", &routers)
	if err != nil {
		return nil, err
	}

	for _, router := range routers {
		router.aisData = aisdata
		router.Decoded = decoded
		router.Failed = failed
		router.inStrings = make(chan string)
	}

	return routers, nil
}

func (router *RemoteAISServer) Start() {
	router.running = true
	go aislib.Router(router.inStrings, router.Decoded, router.Failed)

	timeoutSleep := time.Duration(connRetryTimeoutSecs) * time.Second
	go func() {
		for router.running {
			serverAddr, err := net.ResolveTCPAddr("tcp", router.HostColonPort)
			if err != nil {
				logger.WithError(err).Warnf("could not resolve AIS host %s, retrying in %d secs", router.HostColonPort, connRetryTimeoutSecs)
				router.connAttempts++
				if router.connAttempts > connRetryAttempts {
					logger.Errorf("failing this AIS server %s", router.HostColonPort)
					router.Stop()
					return
				}
				time.Sleep(timeoutSleep)
				continue
			}
			logger.Infof("Resolved host %s", router.HostColonPort)

			router.conn, err = net.DialTCP("tcp", nil, serverAddr)
			if err != nil {
				logger.WithError(err).Warnf("could not connect to AIS host %s, retrying in %d secs", router.HostColonPort, connRetryTimeoutSecs)
				router.connAttempts++
				if router.connAttempts > connRetryAttempts {
					logger.Errorf("failing this AIS server %s", router.HostColonPort)
					router.Stop()
					return
				}
				time.Sleep(timeoutSleep)
				continue
			}
			logger.Infof("Dialed host %+v", router.HostColonPort)

			router.connAttempts = 0

			connbuf := bufio.NewScanner(router.conn)
			connbuf.Split(bufio.ScanLines)
			for connbuf.Scan() && router.running {
				router.inStrings <- connbuf.Text()
			}

			if err := router.conn.Close(); err != nil {
				logger.WithError(err).Error("while closing router")
			}
			logger.Warnf("connection broken/not established to host %s, retrying in %d secs", router.HostColonPort, connRetryTimeoutSecs)
			time.Sleep(timeoutSleep)
		}
		logger.Info("router reconnect loop exiting")
	}()
}

func (router *RemoteAISServer) Stop() {
	router.running = false
}

func (router *RemoteAISServer) DecodePositions(decoded chan aislib.Message, failed chan aislib.FailedSentence) {
	logger.Infof("Starting AIS loop, source %s", router.SourceName)
	for {
		select {
		case message := <-decoded:
			switch message.Type {
			case 1, 2, 3:
				t, err := aislib.DecodeClassAPositionReport(message.Payload)
				if err != nil {
					logger.WithError(err).Warn("decoding class A report")
					break
				}
				report := &SourcedClassAPositionReport{t, SourceAndTime{router.SourceName, time.Now()}}
				logger.Debugf("New type A position '%+v'", report)
				router.aisData.AddPosition(report)

			case 4:
				t, err := aislib.DecodeBaseStationReport(message.Payload)
				if err != nil {
					logger.WithError(err).Warn("decoding base station report")
					break
				}
				report := &SourcedBaseStationReport{t, SourceAndTime{router.SourceName, time.Now()}}
				logger.Debugf("New base station data '%+v'", report)
				router.aisData.UpdateBaseStationReport(report)

			case 5:
				t, err := aislib.DecodeStaticVoyageData(message.Payload)
				if err != nil {
					logger.WithError(err).Warn("decoding voyage data")
					break
				}
				report := &SourcedStaticVoyageData{t, SourceAndTime{router.SourceName, time.Now()}}
				logger.Debugf("New voyage data '%+v'", report)
				router.aisData.UpdateStaticVoyageData(report)

			case 8:
				t, err := aislib.DecodeBinaryBroadcast(message.Payload)
				if err != nil {
					logger.WithError(err).Warn("decoding binary broadcast")
					break
				}
				report := &SourcedBinaryBroadcast{t, SourceAndTime{router.SourceName, time.Now()}}
				logger.Debugf("New binary broadcast '%+v'", report)
				router.aisData.UpdateBinaryBroadcast(report)

			case 18:
				t, err := aislib.DecodeClassBPositionReport(message.Payload)
				if err != nil {
					logger.WithError(err).Warn("decoding class B report")
					break
				}
				report := &SourcedClassBPositionReport{t, SourceAndTime{router.SourceName, time.Now()}}
				logger.Debugf("New type B position '%+v'", report)
				router.aisData.AddPosition(report)

			default:
				logger.Debugf("Unsupported message type %d", message.Type)
			}

		case problematic := <-failed:
			logger.Debugf("Failed message, issue %s, sentence %s", problematic.Issue, problematic.Sentence)
		}
	}
}

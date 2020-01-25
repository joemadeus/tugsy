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
	Source() string
	ReceivedTime() time.Time
}

type SourceAndTime struct {
	sourceName   string
	receivedTime time.Time
}

func (st *SourceAndTime) Source() string {
	return st.sourceName
}

func (st *SourceAndTime) ReceivedTime() time.Time {
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

var MidIso = map[int]string{
	201: "AL",
	202: "AD",
	203: "AT",
	204: "_azores",
	205: "BE",
	206: "BY",
	207: "BG",
	208: "VA",
	209: "CY",
	210: "CY",
	211: "DE",
	212: "CY",
	213: "GE",
	214: "MD",
	215: "MT",
	216: "AM",
	218: "DE",
	219: "DK",
	220: "DK",
	224: "ES",
	225: "ES",
	226: "FR",
	227: "FR",
	228: "FR",
	229: "MT",
	230: "FI",
	231: "FO",
	232: "GB",
	233: "GB",
	234: "GB",
	235: "GB",
	236: "GI",
	237: "GR",
	238: "HR",
	239: "GR",
	240: "GR",
	241: "GR",
	242: "MA",
	243: "HU",
	244: "NL",
	245: "NL",
	246: "NL",
	247: "IT",
	248: "MT",
	249: "MT",
	250: "IE",
	251: "ISO",
	252: "LI",
	253: "LU",
	254: "MC",
	255: "_madeira",
	256: "MT",
	257: "NO",
	258: "NO",
	259: "NO",
	261: "PL",
	262: "ME",
	263: "PT",
	264: "RO",
	265: "SE",
	266: "SE",
	267: "SK",
	268: "SM",
	269: "CH",
	270: "CZ",
	271: "TR",
	272: "UA",
	273: "RU",
	274: "MK",
	275: "LV",
	276: "EE",
	277: "LT",
	278: "SI",
	279: "RS",
	301: "AI",
	303: "US",
	304: "AG",
	305: "AG",
	306: "NL", // cheating: there are three Caribbean nations with designation 306, all formerly Netherland colonies
	307: "AW",
	308: "BS",
	309: "BS",
	310: "BM",
	311: "BS",
	312: "BZ",
	314: "BB",
	316: "CA",
	319: "KY",
	321: "CR",
	323: "CU",
	325: "DM",
	327: "DO",
	329: "GP",
	330: "GD",
	331: "GL",
	332: "GT",
	334: "HN",
	336: "HT",
	338: "US",
	339: "JM",
	341: "KN",
	343: "LC",
	345: "MX",
	347: "MQ",
	348: "MS",
	350: "NI",
	351: "PA",
	352: "PA",
	353: "PA",
	354: "PA",
	355: "PA",
	356: "PA",
	357: "PA",
	358: "PR",
	359: "SV",
	361: "PM",
	362: "TT",
	364: "TC",
	366: "US",
	367: "US",
	368: "US",
	369: "US",
	370: "PA",
	371: "PA",
	372: "PA",
	373: "PA",
	374: "PA",
	375: "VC",
	376: "VC",
	377: "VC",
	378: "VG",
	379: "VI",
	401: "AF",
	403: "SA",
	405: "BD",
	408: "BH",
	410: "BT",
	412: "CN",
	413: "CN",
	414: "CN",
	416: "TW",
	417: "LK",
	419: "IN",
	422: "IR",
	423: "AZ",
	425: "IQ",
	428: "IL",
	431: "JP",
	432: "JP",
	434: "TM",
	436: "KZ",
	437: "UZ",
	438: "JO",
	440: "KR",
	441: "KR",
	443: "_palestine",
	445: "KP",
	447: "KW",
	450: "LB",
	451: "KG",
	453: "MO",
	455: "MV",
	457: "MN",
	459: "NP",
	461: "OM",
	463: "PK",
	466: "QA",
	468: "SY",
	470: "AE",
	471: "AE",
	472: "TJ",
	473: "YE",
	475: "YE",
	477: "HK",
	478: "BA",
	501: "_adelie",
	503: "AU",
	506: "MM",
	508: "BN",
	510: "FM",
	511: "PW",
	512: "NZ",
	514: "KH",
	515: "KH",
	516: "CX",
	518: "CK",
	520: "FJ",
	523: "CC",
	525: "ID",
	529: "KI",
	531: "LA",
	533: "MY",
	536: "MP",
	538: "MH",
	540: "NC",
	542: "NU",
	544: "NR",
	546: "PF",
	548: "PH",
	550: "TL",
	553: "PG",
	555: "PN",
	557: "SB",
	559: "AS",
	561: "WS",
	563: "SG",
	564: "SG",
	565: "SG",
	566: "SG",
	567: "TH",
	570: "TO",
	572: "TV",
	574: "VN",
	576: "VU",
	577: "VU",
	578: "WF",
	601: "ZA",
	603: "AO",
	605: "DZ",
	607: "_stpaul",
	608: "SH",
	609: "BI",
	610: "BJ",
	611: "BW",
	612: "CF",
	613: "CM",
	615: "CG",
	616: "KM",
	617: "CV",
	618: "_crozet",
	619: "CI",
	620: "KM",
	621: "DJ",
	622: "EG",
	624: "ET",
	625: "ER",
	626: "GA",
	627: "GH",
	629: "GM",
	630: "GW",
	631: "GQ",
	632: "GN",
	633: "BF",
	634: "KE",
	635: "_kerguelen",
	636: "LR",
	637: "LR",
	638: "SS",
	642: "LY",
	644: "LS",
	645: "MU",
	647: "MG",
	649: "ML",
	650: "MZ",
	654: "MR",
	655: "MW",
	656: "NE",
	657: "NG",
	659: "NA",
	660: "_reunion",
	661: "RW",
	662: "SD",
	663: "SN",
	664: "SC",
	665: "SH",
	666: "SO",
	667: "SL",
	668: "ST",
	669: "SZ",
	670: "TD",
	671: "TG",
	672: "TN",
	674: "TZ",
	675: "UG",
	676: "CD",
	677: "TZ",
	678: "ZM",
	679: "ZW",
	701: "AR",
	710: "BR",
	720: "BO",
	725: "CL",
	730: "CO",
	735: "EC",
	740: "FK",
	745: "GF",
	750: "GY",
	755: "PY",
	760: "PE",
	765: "SR",
	770: "UY",
	775: "VE",
}

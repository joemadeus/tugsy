package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/vmihailenco/msgpack"
)

const (
	serverport             = ":8080"
	sensorDataEndpointPath = "/sensordata"

	databaseName = "sensordata.sqlite3"
	deployDir    = "/deploy/boatsie"
	devDir       = "."
	resourcesDir = "/resources"

	getSinceTimestampRequestVar = "since"
	getSinceTimestampDefault    = 1 * time.Hour

	getSensorNameRequestVar = "sensor"
	getLocationRequestVar   = "location"
)

// SensorMeta is the metadata for a sensor reading: the name of the sensor, its
// location, when it was recorded at the location and when it was received here
//type SensorMeta struct {
//	Name     string `db:"sensor_name"`
//	Location string `db:"sensor_location"`
//	Recorded int64  `db:"recorded"`
//	Received int64  `db:"received"`
//}

// SensorReading is the metadata for a reading along with K/V pairs specific to
// the sensor that contained the actual recorded data
//type SensorReading struct {
//	SensorMeta
//	Data map[string]interface{}
//}

// A generic error struct for when bad things happen
type Dammit struct {
	ErrorString string `json:"error"`
}

// Returns the resources dir for the app, preferring the OSX dir over the dev dir
func getResourcesDir() string {
	if _, err := os.Stat(deployDir + resourcesDir); err == nil {
		return deployDir + resourcesDir
	} else if _, err := os.Stat(devDir + resourcesDir); err == nil {
		return devDir + resourcesDir
	} else {
		logger.Fatal("Could not determine the base dir for the app")
		return ""
	}
}

type Endpoint struct {
	SensorDataAccessor
	RequestPath   string
	RequestMethod string
	ContentType   string
}

type AddDataEndpoint Endpoint
type GetDataEndpoint Endpoint

func (endpoint *AddDataEndpoint) Runnit(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	w.Header().Set("Content-Type", endpoint.ContentType)
	var err error

	readings := []SensorReading{}
	err = msgpack.NewDecoder(r.Body).Decode(&readings)
	if err != nil {
		logger.Warn("could not decode sensor data", err)
		w.WriteHeader(500)
		msgpack.NewEncoder(w).Encode(Dammit{"could not decode sensor data: " + err.Error()})
		return
	}

	for _, reading := range readings {
		reading.Received = now.Unix()
		err = endpoint.AddSensorJSON(reading)
		if err != nil {
			logger.Warn("could not write sensor data to the DB", err)
			w.WriteHeader(500)
			msgpack.NewEncoder(w).Encode(Dammit{"could not write sensor data to the DB: " + err.Error()})
			return
		}
		logger.Trace("added", "reading", reading)
	}

	w.WriteHeader(202)
	w.Write([]byte("[]\n")) // no response body, really
}

func (endpoint *GetDataEndpoint) Runnit(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	w.Header().Set("Content-Type", endpoint.ContentType)
	var err error

	val := r.FormValue(getSinceTimestampRequestVar)
	var since int64
	if val != "" {
		logger.Debug("parsing a since param")
		since, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			logger.Warn("Couldn't decode a 'since' string", "since", since, "err", err)
		}
	}
	if since == 0 {
		logger.Debug("using default 'since'")
		since = now.Add(-1 * getSinceTimestampDefault).Unix()
	}

	sensor := r.FormValue(getSensorNameRequestVar)
	location := r.FormValue(getLocationRequestVar)

	logger.Debug("Query params", "since", since, "sensor", sensor, "location", location)

	data, err := endpoint.GetSensorJSON(since, sensor, location)
	if err != nil {
		logger.Warn("could not get sensor data from the DB", err)
		w.WriteHeader(500)
		msgpack.NewEncoder(w).Encode(Dammit{"could not get sensor data from the DB: " + err.Error()})
		return
	}

	logger.Debug("returning data", "count", len(data))

	w.WriteHeader(200)
	msgpack.NewEncoder(w).Encode(data)
}

func main() {
	sensorDataAccess, err := NewSensorDataAccess(getResourcesDir() + "/" + databaseName)
	if err != nil {
		logger.Fatal("Couldn't open the database", err)
		return
	}
	defer sensorDataAccess.Close()

	getDataEndpoint := GetDataEndpoint{
		SensorDataAccessor: sensorDataAccess,
		RequestPath:        sensorDataEndpointPath,
		RequestMethod:      "GET",
		ContentType:        "application/json",
	}

	addDataEndpoint := AddDataEndpoint{
		SensorDataAccessor: sensorDataAccess,
		RequestPath:        sensorDataEndpointPath,
		RequestMethod:      "POST",
		ContentType:        "application/json",
	}

	router := mux.NewRouter()
	router.HandleFunc(getDataEndpoint.RequestPath, getDataEndpoint.Runnit).Methods(getDataEndpoint.RequestMethod)
	router.HandleFunc(addDataEndpoint.RequestPath, addDataEndpoint.Runnit).Methods(addDataEndpoint.RequestMethod)
	logger.Info("Server is starting")
	logger.Info("Server is stopping", http.ListenAndServe(serverport, router))
}

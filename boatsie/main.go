package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
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

// SensorReading is the data for a sensor reading: the name of the sensor, its location,
// when it was recorded at the location and when it was received here, along with K/V
// pairs specific to the sensor that contained the actual sensor readings
type SensorReading struct {
	Name     string `json:"name"     db:"sensor_name"`
	Location string `json:"location" db:"sensor_location"`
	Recorded int64  `json:"recorded" db:"recorded"`
	Received int64  `json:"received" db:"received"`

	Data map[string]interface{} `json:"data"`
}

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

type EndpointHandler interface {
	runnit(w http.ResponseWriter, r *http.Request)
}

type Endpoint struct {
	SensorDataAccessor
	RequestPath   string
	RequestMethod string
	ContentType   string
}

type AddDataEndpoint Endpoint
type GetDataEndpoint Endpoint

func (endpoint *AddDataEndpoint) runnit(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	w.Header().Set("Content-Type", endpoint.ContentType)
	var err error

	readings := []SensorReading{}
	err = json.NewDecoder(r.Body).Decode(&readings)
	if err != nil {
		logger.Warn("could not accept sensor data", err)
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(Dammit{"could not accept sensor data: " + err.Error()})
		return
	}

	logger.Trace("Decoded data ", readings)

	for _, reading := range readings {
		reading.Received = now.Unix()
		err = endpoint.AddSensorJSON(reading)
		if err != nil {
			logger.Warn("could not write sensor data to the DB", err)
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(Dammit{"could not write sensor data to the DB: " + err.Error()})
			return
		}
	}

	w.WriteHeader(202)
	w.Write([]byte("[]\n")) // no response body, really
}

func (endpoint *GetDataEndpoint) runnit(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	w.Header().Set("Content-Type", endpoint.ContentType)
	var err error
	var ok bool

	requestVars := mux.Vars(r)
	val, ok := requestVars[getSinceTimestampRequestVar]
	var since int64
	if ok {
		since, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			logger.Warn("Couldn't decode a 'since' string", "since", since, "err", err)
		}
	}
	if since == 0 {
		logger.Debug("using default 'since'")
		since = now.Add(-1 * getSinceTimestampDefault).Unix()
	}

	sensor, ok := requestVars[getSensorNameRequestVar]
	location, ok := requestVars[getLocationRequestVar]

	logger.Debug("Query params", "since", since, "sensor", sensor, "location", location)

	data, err := endpoint.GetSensorJSON(since, sensor, location)
	if err != nil {
		logger.Warn("could not get sensor data from the DB", err)
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(Dammit{"could not get sensor data from the DB: " + err.Error()})
		return
	}

	logger.Debug("returning data", "count", len(data))

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(data)
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
	router.HandleFunc(getDataEndpoint.RequestPath, getDataEndpoint.runnit).Methods(getDataEndpoint.RequestMethod)
	router.HandleFunc(addDataEndpoint.RequestPath, addDataEndpoint.runnit).Methods(addDataEndpoint.RequestMethod)
	logger.Info("Server is starting")
	logger.Info("Server is stopping", http.ListenAndServe(serverport, router))
}

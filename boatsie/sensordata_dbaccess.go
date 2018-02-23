package main

import (
	"encoding/json"

	"github.com/golang/snappy"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const (
    createStatement = `
CREATE TABLE IF NOT EXISTS sensordata (
  sensor_name     TEXT     NOT NULL,
  sensor_location TEXT     NOT NULL,
  recorded        INTEGER  NOT NULL,
  received        INTEGER  NOT NULL,
  json_data       BLOB     NOT NULL,
CONSTRAINT sensordata_pk PRIMARY KEY (sensor_name, sensor_location, recorded))`
)

type SensorDataAccessor interface {
	AddSensorJSON(reading SensorReading) error
	GetSensorJSON(since int64, sensorname string, location string) ([]SensorReading, error)
}

type SensorDataAccess struct {
	*sqlx.DB
}

func NewSensorDataAccess(database string) (*SensorDataAccess, error) {
	db, err := sqlx.Open("sqlite3", database)
	if err != nil {
		return nil, err
	}

	logger.Info("Conditionally creating database")
	tx := db.MustBegin()
	tx.MustExec(createStatement)
	tx.Commit()

	return &SensorDataAccess{db}, nil
}

func (accessor *SensorDataAccess) AddSensorJSON(reading SensorReading) error {
	jsonBytes, err := json.Marshal(reading)
	if err != nil {
		return err
	}

	var compressedBytes []byte
	snappy.Encode(compressedBytes, jsonBytes)

	tx := accessor.DB.MustBegin()
	tx.MustExec(
		"INSERT INTO sensordata (sensor_name, sensor_location, recorded, received, json_data) VALUES ($1, $2, $3, $4, $5)",
		reading.Name, reading.Location, reading.Recorded, reading.Received, compressedBytes)
	tx.Commit()

	return nil
}

func (accessor *SensorDataAccess) GetSensorJSON(since int64, sensorname string, location string) ([]SensorReading, error) {
	selectString := " SELECT json_data FROM sensordata WHERE recorded > $1 "
	sensorString := ""
	if sensorname != "" {
		sensorString = " AND sensor_name = " + sensorname
	}
	locationString := ""
	if location != "" {
		locationString = " AND sensor_location = " + location
	}

	rows := [][]byte{}
	accessor.Select(&rows, selectString+sensorString+locationString)

	readings := make([]SensorReading, 0, len(rows))
	for _, row := range rows {
		var decompressed []byte
		snappy.Decode(row, decompressed)
		reading := SensorReading{}
		err := json.Unmarshal(decompressed, reading)
		if err != nil {
			logger.Warn("could not decode decompressed json", err)
		}
		readings = append(readings, reading)
	}

	return readings, nil
}

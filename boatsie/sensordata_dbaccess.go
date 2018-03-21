package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vmihailenco/msgpack"
)

const (
	createStatement = `
CREATE TABLE IF NOT EXISTS sensordata (
  sensor_name     TEXT     NOT NULL,
  sensor_location TEXT     NOT NULL,
  recorded        INTEGER  NOT NULL,
  received        INTEGER  NOT NULL,
  data            BLOB     NOT NULL,
CONSTRAINT sensordata_pk PRIMARY KEY (sensor_name, sensor_location, recorded))`
)

// The database representation of our sensor readings
type SensorRow struct {
	SensorReading
	Data []byte `db:"data"`
}

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
	dataBytes, err := msgpack.Marshal(reading.Data)
	if err != nil {
		return err
	}

	tx := accessor.DB.MustBegin()
	tx.MustExec(
		"INSERT INTO sensordata (sensor_name, sensor_location, recorded, received, json_data) VALUES ($1, $2, $3, $4, $5)",
		reading.Name, reading.Location, reading.Recorded, reading.Received, dataBytes)
	tx.Commit()

	return nil
}

func (accessor *SensorDataAccess) GetSensorJSON(since int64, sensorname string, location string) ([]SensorReading, error) {
	selectString := " SELECT * FROM sensordata WHERE recorded > $1 " // <- the only bind var! inefficient but I don't care!
	sensorString := ""
	if sensorname != "" {
		sensorString = " AND sensor_name = " + sensorname
	}
	locationString := ""
	if location != "" {
		locationString = " AND sensor_location = " + location
	}

	rows := []SensorRow{}
	accessor.Select(&rows, selectString+sensorString+locationString, since)

	readings := make([]SensorReading, 0, len(rows))
	for _, row := range rows {
		var unmarshalled map[string]interface{}
		err := msgpack.Unmarshal(row.Data, unmarshalled)
		if err != nil {
			logger.Warn("could not decode msgpack'd data map", err)
		}
		reading := SensorReading{
			SensorMeta: row.SensorMeta,
			Data:       unmarshalled,
		}
		readings = append(readings, reading)
	}

	return readings, nil
}

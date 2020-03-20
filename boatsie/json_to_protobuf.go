package main

import (
	"bufio"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
)

type SensorJson struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Recorded int64  `json:"recorded"`
	Received int64  `json:"received,omitempty"`

	Data map[string]interface{} `json:"data"`
}

func main() {
	readings := &SensorReadings{
		Readings: make([]*SensorReading, 0),
	}

	readings.Readings[0] = &SensorReading{
		Name:     "one",
		Location: "top",
		Recorded: time.Now().Unix(),
		Data: []*Data{
			{
				Key:            "a",
				StringOrDouble: &Data_DoubleValue{1.0},
			},
			{
				Key:            "b",
				StringOrDouble: &Data_StringValue{"_b"},
			},
		},
	}

	readings.Readings[1] = &SensorReading{
		Name:     "two",
		Location: "top",
		Recorded: time.Now().Unix(),
		Data: []*Data{
			{
				Key:            "a",
				StringOrDouble: &Data_DoubleValue{10.0},
			},
			{
				Key:            "b",
				StringOrDouble: &Data_StringValue{"_b2"},
			},
		},
	}

	stdout := bufio.NewWriter(os.Stdout)
	bytes, err := proto.Marshal(readings)
	if err != nil {
		panic(err)
	}
	stdout.Write(bytes)
}

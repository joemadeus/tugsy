package main

import (
	"github.com/andmarios/aislib"
)

const masPositionCount = 12

type ShipHistory struct {
	aPositions []SourcedMessage
	bPositions []SourcedMessage
	voyagedata *SourcedMessage
}

type AISData struct {
	maxPositionCount int
	mmsiPositions    map[uint32]ShipHistory
	mmsiBasestations map[uint32]aislib.BaseStationReport
	mmsiBinaryData   map[uint32]aislib.BinaryBroadcast
}

func NewAISData() *AISData {
	return &AISData{
		maxPositionCount: 12,

		mmsiPositions:    make(map[uint32]ShipHistory),
		mmsiBasestations: make(map[uint32]aislib.BaseStationReport),
		mmsiBinaryData:   make(map[uint32]aislib.BinaryBroadcast),
	}
}

func (aisData *AISData) UpdatePosition(report *aislib.PositionReport) {

}

func (aisData *AISData) UpdateVoyageData(data *aislib.StaticVoyageData) {
	aisData.mmsiPositions[data.MMSI] = data
}

func (aisData *AISData) UpdateBaseStationReport(report *aislib.BaseStationReport) {
	aisData.mmsiBasestations[report.MMSI] = report
}

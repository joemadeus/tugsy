package main

import (
	"time"

	"sync"
)

const (
	positionRetentionTime   = time.Duration(12) * time.Hour
	positionCullingInterval = time.Duration(5) * time.Second
	initialPositionCap      = 50
)

type ShipHistory struct {
	aPositions []SourcedClassAPositionReport // TODO: Keep sorted
	bPositions []SourcedClassBPositionReport // TODO: Keep sorted
	voyagedata *SourcedStaticVoyageData
	dirty      bool
	*sync.Mutex
}

func NewShipHistory() ShipHistory {
	return ShipHistory{
		aPositions: make([]SourcedClassAPositionReport, initialPositionCap),
		bPositions: make([]SourcedClassBPositionReport, initialPositionCap),
	}
}

type AISData struct {
	mmsiPositions    map[uint32]ShipHistory
	mmsiBasestations map[uint32]SourcedBaseStationReport
	mmsiBinaryData   map[uint32]SourcedBinaryBroadcast
	*sync.Mutex
}

func NewAISData() *AISData {
	return &AISData{
		mmsiPositions:    make(map[uint32]ShipHistory),
		mmsiBasestations: make(map[uint32]SourcedBaseStationReport),
		mmsiBinaryData:   make(map[uint32]SourcedBinaryBroadcast),
	}
}

func (aisData *AISData) AddAPosition(report *SourcedClassAPositionReport) {
	history := aisData.getOrCreateShipHistory(report.MMSI)
	history.Lock()
	defer history.Unlock()
	history.aPositions = append(history.aPositions, *report)
	history.dirty = true
}

func (aisData *AISData) AddBPosition(report *SourcedClassBPositionReport) {
	history := aisData.getOrCreateShipHistory(report.MMSI)
	history.Lock()
	defer history.Unlock()
	history.bPositions = append(history.bPositions, *report)
	history.dirty = true
}

func (aisData *AISData) PrunePositions() {
	// Create the timer that'll remove old positions
	cullChan := time.Tick(positionCullingInterval)

	for {
		select {
		case <-cullChan:
			logger.Debug("Culling positions")
			// make a copy of the keyset so we don't have to maintain the lock on aisData.
			// this code assumes we accumulate but never dispose of ShipHistory instances
			aisData.Lock()
			mmsis := make([]uint32, len(aisData.mmsiPositions))
			i := 0
			for mmsi := range aisData.mmsiPositions {
				mmsis[i] = mmsi
				i++
			}
			aisData.Unlock()

			for _, mmsi := range mmsis {
				history, ok := aisData.mmsiPositions[mmsi]
				if ok == false {
					logger.Warn("Generated a list of MMSIs from the ais data, but couldn't find an MMSI!", "MMSI", mmsi)
					continue
				}

				history.Lock()
				dirty := false
				// TODO
				history.dirty = dirty
				history.Unlock()
			}
		}
	}
}

// Returns the ShipHistory associated with the given MMSI, or nil if it doesn't exist
func (aisData *AISData) GetShipHistory(mmsi uint32) (*ShipHistory, bool) {
	aisData.Lock()
	defer aisData.Unlock()
	history, ok := aisData.mmsiPositions[mmsi]
	if ok {
		return &history, ok
	} else {
		return nil, ok
	}
}

func (aisData *AISData) getOrCreateShipHistory(mmsi uint32) *ShipHistory {
	aisData.Lock()
	defer aisData.Unlock()
	history, ok := aisData.mmsiPositions[mmsi]
	if ok == false {
		history = NewShipHistory()
		aisData.mmsiPositions[mmsi] = history
	}

	return &history
}

func (aisData *AISData) UpdateStaticVoyageData(data *SourcedStaticVoyageData) {
	aisData.Lock()
	defer aisData.Unlock()
	history := aisData.mmsiPositions[data.MMSI]
	history.voyagedata = data
}

func (aisData *AISData) UpdateBaseStationReport(report *SourcedBaseStationReport) {
	aisData.Lock()
	defer aisData.Unlock()
	aisData.mmsiBasestations[report.MMSI] = *report
}

func (aisData *AISData) UpdateBinaryBroadcast(report *SourcedBinaryBroadcast) {
	aisData.Lock()
	defer aisData.Unlock()
	aisData.mmsiBinaryData[report.MMSI] = *report
}

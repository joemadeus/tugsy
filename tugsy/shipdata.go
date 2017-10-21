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
	// position history is sorted oldest-first -- we append new positions
	// to the end
	aPositions []*SourcedClassAPositionReport // TODO: Keep sorted
	bPositions []*SourcedClassBPositionReport // TODO: Keep sorted
	voyagedata *SourcedStaticVoyageData
	dirty      bool
	*sync.Mutex
}

func NewShipHistory() ShipHistory {
	return ShipHistory{
		aPositions: make([]*SourcedClassAPositionReport, initialPositionCap),
		bPositions: make([]*SourcedClassBPositionReport, initialPositionCap),
	}
}

func (history *ShipHistory) addAndPruneAPosition(report *SourcedClassAPositionReport) {
	// mutex held by calling code
	var newSlice []*SourcedClassAPositionReport
	newSlice = history.prune(history.aPositions)
	newSlice = append(newSlice, report)
	history.dirty = true
	history.aPositions = newSlice
}

func (history *ShipHistory) addAndPruneBPosition(report *SourcedClassBPositionReport) {
	// mutex held by calling code
	var newSlice []*SourcedClassBPositionReport
	newSlice = history.prune(history.bPositions)
	newSlice = append(newSlice, report)
	history.dirty = true
	history.bPositions = newSlice
}

func (history *ShipHistory) prune(reports []SourcedReport) []SourcedReport {
	// mutex held by calling code
	since := time.Now().Add(positionRetentionTime * -1)

	// shortcut -- test the first element. if it's after 'since', just return the original
	// slice and don't flag 'dirty'
	if reports[0].GetReceivedTime().After(since) {
		return reports
	}

	history.dirty = true
	var newSlice []SourcedReport
	i := 0
	for i < len(reports) {
		if reports[i].GetReceivedTime().After(since) {
			// the rest of these should be > since, so just copy them without testing
			newSlice = append(newSlice, reports[i:]...)
			return newSlice
		}
		i++
	}

	return newSlice
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
	history.addAndPruneAPosition(report)
}

func (aisData *AISData) AddBPosition(report *SourcedClassBPositionReport) {
	history := aisData.getOrCreateShipHistory(report.MMSI)
	history.Lock()
	defer history.Unlock()
	history.addAndPruneBPosition(report)
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

// Run as a go func. Periodically, and forever, prune positions from all the known ship
// histories.
func (aisData *AISData) PrunePositions() {
	cullChan := time.Tick(positionCullingInterval)

	for {
		select {
		case <-cullChan:
			logger.Debug("Culling positions")
			// make a copy of the keyset so we don't have to maintain the lock on aisData.
			// doing so means potentially examining only a subset of all the shipdata, but
			// that's alright: this isn't toooo important a process & we'll get to the ones
			// we miss next time
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
					logger.Debug("An MMSI was removed before we got could prune it", "MMSI", mmsi)
					continue
				}

				history.Lock()
				history.aPositions = history.prune(history.aPositions)
				history.bPositions = history.prune(history.bPositions)
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

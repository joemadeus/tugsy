package main

import (
	"time"

	"sync"

	"github.com/andmarios/aislib"
)

const (
	initialPositionCap = 50
)

type ShipHistory struct {
	// position history is sorted oldest-first -- we append new positions to the end
	positions  []Positionable
	voyagedata *SourcedStaticVoyageData
	dirty      bool
	*sync.Mutex
}

func NewShipHistory() ShipHistory {
	return ShipHistory{
		positions: make([]Positionable, 0, initialPositionCap),
	}
}

func (history *ShipHistory) addPosition(report Positionable) {
	// mutex held by calling code
	history.positions = append(history.positions, report)
	history.dirty = true
}

func (history *ShipHistory) prune(since time.Time) {
	// mutex held by calling code

	// shortcut -- test the first element. if it's after 'since', just return the original
	// slice and don't flag 'dirty'
	if len(history.positions) == 0 || history.positions[0].GetReceivedTime().After(since) {
		return
	}

	history.dirty = true

	var a int
	var position Positionable
	for a, position = range history.positions {
		if position.GetReceivedTime().After(since) {
			break
		}
	}

	var newSlice []Positionable
	history.positions = append(newSlice, history.positions[a:]...)
}

type AISData struct {
	mmsiHistories    map[uint32]ShipHistory
	mmsiBasestations map[uint32]SourcedBaseStationReport
	mmsiBinaryData   map[uint32]SourcedBinaryBroadcast

	positionRetentionTime   time.Duration
	positionCullingInterval time.Duration

	dirty bool
	sync.Mutex
}

func NewAISData() *AISData {
	return &AISData{
		mmsiHistories:    make(map[uint32]ShipHistory),
		mmsiBasestations: make(map[uint32]SourcedBaseStationReport),
		mmsiBinaryData:   make(map[uint32]SourcedBinaryBroadcast),

		positionRetentionTime:   time.Duration(12) * time.Hour,
		positionCullingInterval: time.Duration(5) * time.Second,
	}
}

func (aisData *AISData) AddPosition(report Positionable) {
	history := aisData.getOrCreateShipHistory(report.GetPositionReport().MMSI)
	history.Lock()
	defer history.Unlock()
	history.addPosition(report)
	history.prune(aisData.getPruneSinceTime())
	aisData.dirty = true
}

func (aisData *AISData) getOrCreateShipHistory(mmsi uint32) *ShipHistory {
	aisData.Lock()
	defer aisData.Unlock()
	history, ok := aisData.mmsiHistories[mmsi]
	if ok == false {
		history = NewShipHistory()
		aisData.mmsiHistories[mmsi] = history
	}

	return &history
}

func (aisData *AISData) UpdateStaticVoyageData(data *SourcedStaticVoyageData) {
	aisData.Lock()
	defer aisData.Unlock()
	history := aisData.mmsiHistories[data.MMSI]
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
	cullChan := time.Tick(aisData.positionCullingInterval)

	for {
		select {
		case <-cullChan:
			logger.Debug("Culling positions")
			// make a copy of the keyset so we don't have to maintain the lock on aisData.
			// doing so means potentially examining only a subset of all the shipdata, but
			// that's alright: this isn't toooo important a process & we'll get to the ones
			// we miss next time
			since := aisData.getPruneSinceTime()
			for _, mmsi := range aisData.GetHistoryMMSIs() {
				history, ok := aisData.mmsiHistories[mmsi]
				if ok == false {
					logger.Debug("An MMSI was removed before we got could prune it", "MMSI", mmsi)
					continue
				}

				history.Lock()
				history.prune(since)
				if history.dirty {
					aisData.dirty = true
				}
				history.Unlock()
			}
		}
	}
}

// Returns a slice of all known MMSI values
func (aisData *AISData) GetHistoryMMSIs() []uint32 {
	aisData.Lock()
	defer aisData.Unlock()
	mmsis := make([]uint32, 0, len(aisData.mmsiHistories))
	i := 0
	for mmsi := range aisData.mmsiHistories {
		mmsis[i] = mmsi
		i++
	}

	return mmsis
}

// Returns the ShipHistory/true associated with the given MMSI, or nil/false if it doesn't.
// Calling code should lock using the history's mutex if modifying or querying data.
func (aisData *AISData) GetShipHistory(mmsi uint32) (*ShipHistory, bool) {
	aisData.Lock()
	defer aisData.Unlock()
	history, ok := aisData.mmsiHistories[mmsi]
	if ok {
		return &history, ok
	} else {
		return nil, ok
	}
}

// Returns a slice with all the position reports, sorted by time received ascending, or nil
// if the given MMSI is unknown. If the history is known, sets the given dirty val on the history
func (aisData *AISData) GetPositionReports(mmsi uint32, newDirtyVal bool) []*aislib.PositionReport {
	aisData.Lock()
	history, ok := aisData.GetShipHistory(mmsi)
	aisData.Unlock()

	if ok == false {
		return nil
	}

	history.Lock()
	defer history.Unlock()

	positionCount := len(history.positions)
	positions := make([]*aislib.PositionReport, 0, positionCount)
	for i := 0; i < positionCount; i++ {
		positions[i] = history.positions[i].GetPositionReport()
	}

	history.dirty = newDirtyVal
	return positions
}

func (aisData *AISData) getPruneSinceTime() time.Time {
	return time.Now().Add(aisData.positionRetentionTime * -1)
}

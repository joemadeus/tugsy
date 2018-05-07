package shipdata

import (
	"sync"
	"time"
)

const (
	defaultPositionRetentionTime   = 12 * time.Hour
	defaultPositionCullingInterval = 5 * time.Second
)

type ShipHistory struct {
	// append new positions to the end
	positions  []Positionable
	voyagedata *SourcedStaticVoyageData
	dirty      bool
	sync.Mutex
}

func NewShipHistory() *ShipHistory {
	return &ShipHistory{positions: make([]Positionable, 0)}
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

var PositionData = &AISData{
	mmsiHistories:    make(map[uint32]*ShipHistory),
	mmsiBaseStations: make(map[uint32]*SourcedBaseStationReport),
	mmsiBinaryData:   make(map[uint32]*SourcedBinaryBroadcast),

	positionRetentionTime:   defaultPositionRetentionTime,
	positionCullingInterval: defaultPositionCullingInterval,
}

type AISData struct {
	mmsiHistories    map[uint32]*ShipHistory
	mmsiBaseStations map[uint32]*SourcedBaseStationReport
	mmsiBinaryData   map[uint32]*SourcedBinaryBroadcast

	positionRetentionTime   time.Duration
	positionCullingInterval time.Duration

	sync.Mutex
}

func (aisData *AISData) AddPosition(report Positionable) {
	aisData.Lock()
	history := aisData.getOrCreateShipHistory(report.GetPositionReport().MMSI)
	aisData.Unlock()

	history.Lock()
	defer history.Unlock()
	history.addPosition(report)
	history.prune(aisData.getPruneSinceTime())
}

func (aisData *AISData) getOrCreateShipHistory(mmsi uint32) *ShipHistory {
	// mutex held by calling code
	history, ok := aisData.mmsiHistories[mmsi]
	if ok == false {
		history = NewShipHistory()
		aisData.mmsiHistories[mmsi] = history
	}

	return history
}

func (aisData *AISData) UpdateStaticVoyageData(data *SourcedStaticVoyageData) {
	aisData.Lock()
	history := aisData.getOrCreateShipHistory(data.MMSI)
	aisData.Unlock()

	history.Lock()
	defer history.Unlock()
	history.voyagedata = data
}

func (aisData *AISData) UpdateBaseStationReport(report *SourcedBaseStationReport) {
	aisData.Lock()
	defer aisData.Unlock()
	aisData.mmsiBaseStations[report.MMSI] = report
}

func (aisData *AISData) UpdateBinaryBroadcast(report *SourcedBinaryBroadcast) {
	aisData.Lock()
	defer aisData.Unlock()
	aisData.mmsiBinaryData[report.MMSI] = report
}

// Run as a go func. Periodically, and forever, prune positions from all the known ship
// histories.
func (aisData *AISData) PrunePositions() {
	cullChan := time.Tick(aisData.positionCullingInterval)

	for {
		select {
		case <-cullChan:
			logger.Trace("Culling positions")
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

				if len(history.positions) == 0 {
					logger.Info("A ship has not been heard from in a while. Removing.", "mmsi", mmsi)
					delete(aisData.mmsiHistories, mmsi)
				}

				history.Unlock()
			}
		}
	}
}

// Returns a slice of all known ShipHistory MMSI values
func (aisData *AISData) GetHistoryMMSIs() []uint32 {
	aisData.Lock()
	defer aisData.Unlock()
	mmsis := make([]uint32, len(aisData.mmsiHistories))
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
		return history, ok
	} else {
		return nil, ok
	}
}

func (aisData *AISData) getPruneSinceTime() time.Time {
	return time.Now().Add(aisData.positionRetentionTime * -1)
}

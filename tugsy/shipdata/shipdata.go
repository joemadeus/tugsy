package shipdata

import (
	"sync"
	"time"

	logger "github.com/sirupsen/logrus"
)

const (
	defaultPositionRetentionDur    = 12 * time.Hour
	defaultPositionCullingInterval = 5 * time.Second
)

type ShipHistory struct {
	sync.Mutex

	// append new positions to the end
	positions  []Positionable
	voyagedata *SourcedStaticVoyageData
	dirty      bool
}

func NewShipHistory() *ShipHistory {
	return &ShipHistory{positions: make([]Positionable, 0)}
}

func (h *ShipHistory) addPosition(report Positionable) {
	h.Lock()
	defer h.Unlock()

	h.positions = append(h.positions, report)
	h.dirty = true
}

func (h *ShipHistory) setVoyageData(d *SourcedStaticVoyageData) {
	h.Lock()
	defer h.Unlock()

	h.voyagedata = d
}

func (h *ShipHistory) prune(since time.Time) {
	h.Lock()
	defer h.Unlock()

	// shortcut -- test the first element. if it's after 'since', just return the original
	// slice and don't flag 'dirty'
	if len(h.positions) == 0 || h.positions[0].GetReceivedTime().After(since) {
		return
	}

	h.dirty = true

	var a int
	var position Positionable
	for a, position = range h.positions {
		if position.GetReceivedTime().After(since) {
			break
		}
	}

	h.positions = h.positions[a:]
}

type AISData struct {
	sync.Mutex

	mmsiHistories    map[uint32]*ShipHistory
	mmsiBaseStations map[uint32]*SourcedBaseStationReport
	mmsiBinaryData   map[uint32]*SourcedBinaryBroadcast

	positionRetentionDur    time.Duration
	positionCullingInterval time.Duration
}

func NewAISData() *AISData {
	return &AISData{
		mmsiHistories:    make(map[uint32]*ShipHistory),
		mmsiBaseStations: make(map[uint32]*SourcedBaseStationReport),
		mmsiBinaryData:   make(map[uint32]*SourcedBinaryBroadcast),

		positionRetentionDur:    defaultPositionRetentionDur,
		positionCullingInterval: defaultPositionCullingInterval,
	}
}

func (aisData *AISData) AddPosition(report Positionable) {
	history := aisData.getOrCreateShipHistory(report.GetPositionReport().MMSI)
	history.addPosition(report)
}

func (aisData *AISData) getOrCreateShipHistory(mmsi uint32) *ShipHistory {
	aisData.Lock()
	defer aisData.Unlock()

	history, ok := aisData.mmsiHistories[mmsi]
	if ok == false {
		history = NewShipHistory()
		aisData.mmsiHistories[mmsi] = history
	}

	return history
}

func (aisData *AISData) UpdateStaticVoyageData(data *SourcedStaticVoyageData) {
	history := aisData.getOrCreateShipHistory(data.MMSI)
	history.setVoyageData(data)
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
			logger.Debug("culling positions")
			// make a copy of the keyset so we don't have to maintain the lock on aisData.
			// doing so means potentially examining only a subset of all the shipdata, but
			// that's alright: this isn't toooo important a process & we'll get to the ones
			// we miss next time
			since := time.Now().Add(-aisData.positionRetentionDur)
			for _, mmsi := range aisData.GetHistoryMMSIs() {
				history, ok := aisData.mmsiHistories[mmsi]
				if ok == false {
					logger.Debugf("MMSI %d was removed before it could be pruned", mmsi)
					continue
				}

				history.Lock()
				history.prune(since)

				if len(history.positions) == 0 {
					logger.Infof("a ship has not been heard from in a while. Removing MMSI %v", mmsi)
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

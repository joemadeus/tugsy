package shipdata

import (
	"fmt"
	"sync"
	"time"

	logger "github.com/sirupsen/logrus"
)

const (
	defaultPositionRetentionDur    = 12 * time.Hour
	defaultPositionCullingInterval = 5 * time.Second
)

type MMSIUnavailableError struct {
	MMSI uint32
}

func (e MMSIUnavailableError) Error() string {
	return fmt.Sprintf("mmsi %d is unavailable", e.MMSI)
}

type ShipHistory struct {
	sync.Mutex

	MMSI uint32

	// append new positions to the end
	positions  []Positionable
	voyagedata *SourcedStaticVoyageData
}

func NewShipHistory(mmsi uint32) *ShipHistory {
	return &ShipHistory{MMSI: mmsi, positions: make([]Positionable, 0)}
}

// Positions returns a copy of the slice of Positionables currently associated with
// this ShipHistory
func (h *ShipHistory) Positions() []Positionable {
	h.Lock()
	defer h.Unlock()
	ret := make([]Positionable, len(h.positions), len(h.positions))
	copy(ret, h.positions)
	return ret
}

func (h *ShipHistory) VoyageData() *SourcedStaticVoyageData {
	h.Lock()
	defer h.Unlock()
	return h.voyagedata
}

func (h *ShipHistory) addPosition(report Positionable) {
	h.Lock()
	defer h.Unlock()
	h.positions = append(h.positions, report)
}

func (h *ShipHistory) setVoyageData(d *SourcedStaticVoyageData) {
	h.Lock()
	defer h.Unlock()

	h.voyagedata = d
}

func (h *ShipHistory) prune(since time.Time) int {
	h.Lock()
	defer h.Unlock()

	// shortcut -- test the first element. if it's after 'since', just return the original
	// slice and don't flag 'dirty'
	if len(h.positions) == 0 || h.positions[0].ReceivedTime().After(since) {
		return len(h.positions)
	}

	var a int
	var position Positionable
	for a, position = range h.positions {
		if position.ReceivedTime().After(since) {
			break
		}
	}

	h.positions = h.positions[a:]
	return len(h.positions)
}

type AISData struct {
	sync.Mutex

	mmsiHistories    map[uint32]*ShipHistory
	mmsiBaseStations map[uint32]*SourcedBaseStationReport
	mmsiBinaryData   map[uint32]*SourcedBinaryBroadcast

	PositionRetentionDur    time.Duration
	PositionCullingInterval time.Duration
}

func NewAISData() *AISData {
	return &AISData{
		mmsiHistories:    make(map[uint32]*ShipHistory),
		mmsiBaseStations: make(map[uint32]*SourcedBaseStationReport),
		mmsiBinaryData:   make(map[uint32]*SourcedBinaryBroadcast),

		PositionRetentionDur:    defaultPositionRetentionDur,
		PositionCullingInterval: defaultPositionCullingInterval,
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
		history = NewShipHistory(mmsi)
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

// Forever periodically prunes positions from all the known ship histories.
func (aisData *AISData) PrunePositions() {
	tick := time.Tick(aisData.PositionCullingInterval)

	for {
		select {
		case <-tick:
			logger.Debug("culling positions")
			// make a copy of the keyset so we don't have to maintain the lock on aisData.
			// doing so means potentially examining only a subset of all the shipdata, but
			// that's alright: this isn't toooo important a process & we'll get to the ones
			// we miss next time
			since := time.Now().Add(-aisData.PositionRetentionDur)
			for _, sh := range aisData.ShipHistories() {
				if sh.prune(since) == 0 {
					aisData.Lock()
					// retest for positions within lock
					if len(sh.positions) == 0 {
						logger.Infof("a ship has not been heard from in a while. Removing MMSI %v", sh.MMSI)
						delete(aisData.mmsiHistories, sh.MMSI)
					}
					aisData.Unlock()
				}
			}
		}
	}
}

// Returns a copy of the slice of ShipHistories
func (aisData *AISData) ShipHistories() []*ShipHistory {
	aisData.Lock()
	defer aisData.Unlock()

	shs := make([]*ShipHistory, len(aisData.mmsiHistories), len(aisData.mmsiHistories))
	i := 0
	for _, sh := range aisData.mmsiHistories {
		shs[i] = sh
		i++
	}

	return shs
}

// Returns the ShipHistory/true associated with the given MMSI, or nil/false if it doesn't.
// Calling code should lock using the history's mutex if modifying or querying data.
func (aisData *AISData) ShipHistory(mmsi uint32) (*ShipHistory, bool) {
	aisData.Lock()
	defer aisData.Unlock()
	history, ok := aisData.mmsiHistories[mmsi]
	if ok {
		return history, ok
	} else {
		return nil, ok
	}
}

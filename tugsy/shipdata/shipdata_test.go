package shipdata

import (
	"testing"
	"time"

	"github.com/andmarios/aislib"
	"github.com/stretchr/testify/assert"
)

type MockPositionReport struct {
	receivedTime   time.Time
	positionReport *aislib.PositionReport
}

func NewMockPositionReport(receivedTime time.Time, mmsi uint32) *MockPositionReport {
	return &MockPositionReport{
		receivedTime:   receivedTime,
		positionReport: &aislib.PositionReport{MMSI: mmsi},
	}
}

func (m *MockPositionReport) GetPositionReport() *aislib.PositionReport {
	return m.positionReport
}
func (m *MockPositionReport) GetSource() string {
	return "test"
}
func (m *MockPositionReport) GetReceivedTime() time.Time {
	return m.receivedTime
}

func TestAddAndPrune(t *testing.T) {
	now := time.Now()
	sh := NewShipHistory()
	posOne := &MockPositionReport{receivedTime: now.Add(-10 * time.Second)}
	posTwo := &MockPositionReport{receivedTime: now.Add(10 * time.Second)}

	assert.False(t, sh.dirty)
	sh.addPosition(posOne)
	assert.True(t, sh.dirty)
	sh.addPosition(posTwo)
	assert.True(t, sh.dirty)
	assert.Equal(t, 2, len(sh.positions))
	assert.Equal(t, sh.positions[0], posOne)
	assert.Equal(t, sh.positions[1], posTwo)

	sh.dirty = false
	sh.prune(now)

	assert.True(t, sh.dirty)
	assert.Equal(t, 1, len(sh.positions))
	assert.Equal(t, sh.positions[0], posTwo)
}

func TestNothingToPrune(t *testing.T) {
	now := time.Now()
	sh := NewShipHistory()
	posOne := &MockPositionReport{receivedTime: now.Add(10 * time.Second)}
	posTwo := &MockPositionReport{receivedTime: now.Add(20 * time.Second)}
	sh.addPosition(posOne)
	sh.addPosition(posTwo)
	assert.Equal(t, 2, len(sh.positions))
	assert.True(t, sh.dirty)

	sh.dirty = false
	sh.prune(time.Now())

	assert.Equal(t, 2, len(sh.positions))
	assert.False(t, sh.dirty)
}

func TestEmptyPrune(t *testing.T) {
	sh := NewShipHistory()
	assert.Equal(t, 0, len(sh.positions))
	assert.False(t, sh.dirty)

	sh.prune(time.Now())
	assert.Equal(t, 0, len(sh.positions))
	assert.False(t, sh.dirty)
}

func TestGetPositionReports(t *testing.T) {
	now := time.Now()
	sh := NewShipHistory()
	posOne := &MockPositionReport{receivedTime: now.Add(-10 * time.Second)}
	posTwo := &MockPositionReport{receivedTime: now.Add(10 * time.Second)}

	sh.addPosition(posOne)
	sh.addPosition(posTwo)
	sh.dirty = false
}

func TestTranslatePositionReports(t *testing.T) {
	now := time.Now()
	mmsi := uint32(123)
	aisData := NewAISData()
	aisData.AddPosition(NewMockPositionReport(now, mmsi))
	aisData.AddPosition(NewMockPositionReport(now, mmsi))

	mmsis := aisData.GetHistoryMMSIs()
	assert.Equal(t, 1, len(mmsis))
	assert.Equal(t, mmsi, mmsis[0])
	assert.True(t, aisData.Dirty)

	shipHistory, ok := aisData.GetShipHistory(mmsi)
	assert.True(t, ok)
	assert.Equal(t, 2, len(shipHistory.positions))

	translateOut := make([]uint32, 0, 2)
	translate := func(positionReports []Positionable) {
		for i, val := range positionReports {
			translateOut = append(translateOut, val.GetPositionReport().MMSI*uint32(i+2))
		}
	}

	translateReturn := aisData.TranslatePositionReports(987, translate)
	assert.False(t, translateReturn)

	translateReturn = aisData.TranslatePositionReports(mmsi, translate)
	assert.True(t, translateReturn)
	assert.Equal(t, 2, len(translateOut))
	assert.Equal(t, uint32(246), translateOut[0])
	assert.Equal(t, uint32(369), translateOut[1])
}

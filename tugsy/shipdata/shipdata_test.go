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

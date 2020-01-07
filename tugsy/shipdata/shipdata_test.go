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

func (m *MockPositionReport) Source() string {
	return "test"
}

func (m *MockPositionReport) ReceivedTime() time.Time {
	return m.receivedTime
}

func TestAddAndPrune(t *testing.T) {
	now := time.Now()
	sh := NewShipHistory(1)
	posOne := &MockPositionReport{receivedTime: now.Add(-10 * time.Second)}
	posTwo := &MockPositionReport{receivedTime: now.Add(10 * time.Second)}

	sh.addPosition(posOne)
	sh.addPosition(posTwo)
	assert.Equal(t, 2, len(sh.positions))
	assert.Equal(t, sh.positions[0], posOne)
	assert.Equal(t, sh.positions[1], posTwo)

	sh.prune(now)
	assert.Equal(t, 1, len(sh.positions))
	assert.Equal(t, sh.positions[0], posTwo)
}

func TestNothingToPrune(t *testing.T) {
	now := time.Now()
	sh := NewShipHistory(1)
	posOne := &MockPositionReport{receivedTime: now.Add(10 * time.Second)}
	posTwo := &MockPositionReport{receivedTime: now.Add(20 * time.Second)}
	sh.addPosition(posOne)
	sh.addPosition(posTwo)
	sh.prune(time.Now())
	assert.Equal(t, 2, len(sh.positions))
}

func TestEmptyPrune(t *testing.T) {
	sh := NewShipHistory(1)
	sh.prune(time.Now())
	assert.Equal(t, 0, len(sh.positions))
}

func TestGetPositionReports(t *testing.T) {
	now := time.Now()
	sh := NewShipHistory(1)
	posOne := &MockPositionReport{receivedTime: now.Add(-10 * time.Second)}
	posTwo := &MockPositionReport{receivedTime: now.Add(10 * time.Second)}

	sh.addPosition(posOne)
	sh.addPosition(posTwo)
}

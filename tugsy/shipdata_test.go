package main

import (
	"testing"
	"time"

	"github.com/andmarios/aislib"
	"github.com/stretchr/testify/assert"
)

type MockPositionReport struct {
	receivedTime time.Time
}

func (m *MockPositionReport) GetPositionReport() *aislib.PositionReport {
	return &aislib.PositionReport{}
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

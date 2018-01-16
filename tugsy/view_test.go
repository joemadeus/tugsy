package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBaseMapPosition(t *testing.T) {
	view := View{
		BaseMap: &BaseMap{
			NEGeo:  RealWorldPosition{-60.0, 50.0},
			SWGeo:  RealWorldPosition{-70.0, 40.0},
			width:  100.0,
			height: 200.0,
		},
		ViewName: "test",
	}

	centerUpper := view.getBaseMapPosition(RealWorldPosition{-65.0, 47.5})
	assert.Equal(t, 50.0, centerUpper.X)
	assert.Equal(t, 150.0, centerUpper.Y)

	centerLower := view.getBaseMapPosition(RealWorldPosition{-65.0, 42.5})
	assert.Equal(t, 50.0, centerLower.X)
	assert.Equal(t, 50.0, centerLower.Y)

	upperCenter := view.getBaseMapPosition(RealWorldPosition{-67.5, 45.0})
	assert.Equal(t, 75.0, upperCenter.X) // WEST of origin
	assert.Equal(t, 100.0, upperCenter.Y)

	lowerCenter := view.getBaseMapPosition(RealWorldPosition{-62.5, 45.0})
	assert.Equal(t, 25.0, lowerCenter.X)
	assert.Equal(t, 100.0, lowerCenter.Y)
}
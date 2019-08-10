package views

import (
	"testing"

	"github.com/andmarios/aislib"
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

	centerUpper := view.GetBaseMapPosition(&aislib.PositionReport{Lon: -65.0, Lat: 47.5})
	assert.Equal(t, 50.0, centerUpper.X)
	assert.Equal(t, 50.0, centerUpper.Y)

	centerLower := view.GetBaseMapPosition(&aislib.PositionReport{Lon: -65.0, Lat: 42.5})
	assert.Equal(t, 50.0, centerLower.X)
	assert.Equal(t, 150.0, centerLower.Y)

	upperCenter := view.GetBaseMapPosition(&aislib.PositionReport{Lon: -67.5, Lat: 45.0})
	assert.Equal(t, 25.0, upperCenter.X) // WEST of origin
	assert.Equal(t, 100.0, upperCenter.Y)

	lowerCenter := view.GetBaseMapPosition(&aislib.PositionReport{Lon: -62.5, Lat: 45.0})
	assert.Equal(t, 75.0, lowerCenter.X)
	assert.Equal(t, 100.0, lowerCenter.Y)
}

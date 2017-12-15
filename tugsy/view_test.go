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
	assert.Equal(t, 150.0, centerUpper.Y)
	assert.Equal(t, 50.0, centerUpper.X)
}

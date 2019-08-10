package shipdata

import (
	"errors"
	"fmt"

	"github.com/joemadeus/tugsy/tugsy/views"
)

type ShipInfoElement struct {
	aisData AISData
	MMSI    uint32

	FlagSheet *views.FlagSheet
}

func NewShipInfoElement(spriteSet *views.SpriteSet) (*ShipInfoElement, error) {
	return &ShipInfoElement{
		FlagSheet: spriteSet.FlagSheet,
	}, nil
}

func (style *ShipInfoElement) SetMMSI(mmsi uint32) {
	style.MMSI = mmsi
}

// ShipInfoElement renders information for the currently selected ship, including its
// registration (including flag), name, current destination and current situation
// (moored, underway, etc)
func (style *ShipInfoElement) Render(view *views.View) error {
	if _, ok := style.aisData.GetShipHistory(style.MMSI); ok == false {
		return errors.New(fmt.Sprintf("vessel disappeared before displaying info pane: mmsi %d", style.MMSI))
	}

	// TODO

	return nil
}

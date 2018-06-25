package shipdata

import (
	"errors"
	"fmt"

	"github.com/joemadeus/tugsy/tugsy/views"
)

type ShipInfoStyle struct {
	aisData AISData
	MMSI    uint32

	FlagSheet *views.FlagSheet
}

func NewShipInfoStyle(spriteSet *views.SpriteSet) (*ShipInfoStyle, error) {
	return &ShipInfoStyle{
		FlagSheet: spriteSet.FlagSheet,
	}, nil
}

func (style *ShipInfoStyle) SetMMSI(mmsi uint32) {
	style.MMSI = mmsi
}

// ShipInfoStyle renders information for the currently selected ship, including its
// registration (including flag), name, current destination and current situation
// (moored, underway, etc) into the info pane
func (style *ShipInfoStyle) Render(view *views.View) error {
	if _, ok := style.aisData.GetShipHistory(style.MMSI); ok == false {
		return errors.New(fmt.Sprintf("vessel disappeared before displaying info pane: mmsi %s", style.MMSI))
	}

	// TODO

	return nil
}

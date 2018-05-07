package shipdata

import (
	"errors"
	"fmt"

	"github.com/joemadeus/tugsy/tugsy/views"
)

type ShipInfoStyle struct {
	aisData AISData
	MMSI    uint32

	InfoPane  *views.InfoPaneStyle
	FlagSheet *views.FlagSheet
}

func NewShipInfoStyle(infoPane *views.InfoPaneStyle, flags *views.FlagSheet) *ShipInfoStyle {
	return &ShipInfoStyle{
		InfoPane:  infoPane,
		FlagSheet: flags,
	}
}

func (style *ShipInfoStyle) SetMMSI(mmsi uint32) {
	style.MMSI = mmsi
}

// ShipInfoStyle renders information for the currently selected ship, including its
// registration (including flag), name, current destination and current situation
// (moored, underway, etc) into the info pane
func (style *ShipInfoStyle) Render(view *views.View) error {
	_, ok := style.aisData.GetShipHistory(style.MMSI)
	if ok == false {
		return errors.New(fmt.Sprintf("vessel disappeared before displaying info pane: mmsi %s", style.MMSI))
	}

	err := style.InfoPane.Render(view)
	if err != nil {
		return err
	}

	return nil
}

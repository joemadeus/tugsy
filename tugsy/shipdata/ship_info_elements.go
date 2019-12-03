package shipdata

import (
	"github.com/joemadeus/tugsy/tugsy/views"
)

type ShipInfoElement struct {
	*views.InfoElement
	FlagSheet *views.FlagSheet

	aisData AISData
	mmsi    uint32
}

// ShipInfoElement renders information for the currently selected ship, including its
// registration (including flag), name, current destination and current situation
// (moored, underway, etc) into an InfoElement
func (e *ShipInfoElement) Render(view *views.View) error {
	e.Lock()
	defer e.Unlock()

	if _, ok := e.aisData.GetShipHistory(e.mmsi); ok == false {
		return MMSIUnavailableError{MMSI: e.mmsi}
	}

	// TODO info pane text/flag/etc rendering

	return nil
}

func (e *ShipInfoElement) MMSI(mmsi uint32) {
	e.Lock()
	defer e.Unlock()
	e.mmsi = mmsi
}

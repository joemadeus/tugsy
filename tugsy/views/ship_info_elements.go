package views

import "github.com/joemadeus/tugsy/tugsy/shipdata"

type ShipInfoElement struct {
	*InfoElement
	FlagSheet *FlagSheet

	aisData shipdata.AISData
	mmsi    uint32
}

// ShipInfoElement renders information for the currently selected ship, including its
// registration (including flag), name, current destination and current situation
// (moored, underway, etc) into an InfoElement
func (e *ShipInfoElement) Render(view *View) error {
	e.Lock()
	defer e.Unlock()

	if _, ok := e.aisData.ShipHistory(e.mmsi); ok == false {
		return shipdata.MMSIUnavailableError{MMSI: e.mmsi}
	}

	// TODO info pane text/flag/etc rendering

	return nil
}

func (e *ShipInfoElement) MMSI(mmsi uint32) {
	e.Lock()
	defer e.Unlock()
	e.mmsi = mmsi
}

package shipdata

import "github.com/joemadeus/tugsy/tugsy/views"

type ShipInfoStyle struct{}

// Displays information for the currently selected ship, including its
// registration (including flag), name, current destination and current
// situation (moored, underway, etc)
func (shipinfo *ShipInfoStyle) Render(view *views.View) ShipDataRenderFunction {
	return nil
}

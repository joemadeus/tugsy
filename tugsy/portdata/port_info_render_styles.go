package portdata

import "github.com/joemadeus/tugsy/tugsy/views"

type PortInfoStyle struct {
	InfoPane *views.InfoPaneStyle
}

func NewPortInfoStyle(infoPane *views.InfoPaneStyle) *PortInfoStyle {
	return &PortInfoStyle{InfoPane: infoPane}
}

// PortInfoStyle renders the port name, its code, and a single list of N arrivals and
// departures into the info pane. This is the default display for the info pane
func (style *PortInfoStyle) Render(view *views.View) error {
	style.InfoPane.Render(view)
}

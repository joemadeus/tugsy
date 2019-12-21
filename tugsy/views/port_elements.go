package views

import "github.com/joemadeus/tugsy/tugsy/views"

type PortInfoElement struct {
	CurrentPort string
}

func NewPortInfoElement(port string) *PortInfoElement {
	return &PortInfoElement{CurrentPort: port}
}

// PortInfoElement renders the port name, its code, and a single list of N arrivals and
// departures into the info pane. This is the default display for the info pane
func (style *PortInfoElement) Render(view *views.View) error {
	return nil
}

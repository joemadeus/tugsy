package portdata

import "github.com/joemadeus/tugsy/tugsy/views"

const (
	DefaultPort = "PVD"
)

type PortInfoStyle struct {
	CurrentPort string
}

func NewPortInfoStyle() *PortInfoStyle {
	return &PortInfoStyle{CurrentPort: DefaultPort}
}

// PortInfoStyle renders the port name, its code, and a single list of N arrivals and
// departures into the info pane. This is the default display for the info pane
func (style *PortInfoStyle) Render(view *views.View) error {
	return nil
}

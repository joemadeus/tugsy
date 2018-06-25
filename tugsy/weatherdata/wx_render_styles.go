package weatherdata

import "github.com/joemadeus/tugsy/tugsy/views"

type TideBarStyle struct {
}

// A TideBarStyle renders the current position of the tide, whether it's advancing or
// retreating and its high and low water marks
func (style *TideBarStyle) Render(view *views.View) error {
	return nil
}

// A WxButtonStyle renders a circle colored with the current sky color, plus UV, hazardous
// weather and lightning warning indicators
type StandardWxButtonStyle struct {
}

func (style *StandardWxButtonStyle) Render(view *views.View) error {
	return nil
}

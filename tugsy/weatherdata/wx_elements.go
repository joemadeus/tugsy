package weatherdata

import "github.com/joemadeus/tugsy/tugsy/views"

type TideBarElement struct{}

// A TideBarElement renders the current position of the tide, whether it's advancing or
// retreating and its high and low water marks
func (style *TideBarElement) Render(view *views.View) error {
	return nil
}

// A WxButtonElement renders a circle colored with the current sky color, plus UV, hazardous
// weather and lightning warning indicators
type WxButtonElement struct{}

func (style *WxButtonElement) Render(view *views.View) error {
	return nil
}

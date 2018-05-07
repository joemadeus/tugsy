package weatherdata

import "github.com/joemadeus/tugsy/tugsy/views"

type TideBarStyle views.PaneRenderStyle

// A TideBarStyle renders the current position of the tide, whether it's advancing or
// retreating and its high and low water marks into the weather pane
func (style *TideBarStyle) Render(view *views.View) error {
	return nil
}

// A WxButtonStyle renders a circle colored with the current sky color and UV, hazardous
// weather and lightning warning indicators into the weather pane
type WxButtonStyle views.PaneRenderStyle

func (style *WxButtonStyle) Render(view *views.View) error {
	return nil
}

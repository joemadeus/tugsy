package weatherdata

import "github.com/joemadeus/tugsy/tugsy/views"

type WxDataRenderFunction func(wx *WeatherData)

type TideBarStyle struct{}

// A TideBar displays the current position of the tide, whether it's advancing or
// retreating and its high and low water marks
func (style *TideBarStyle) Render(view *views.View) WxDataRenderFunction {
	return nil
}

// A WxButton is colored with the current sky color and displays UV and hazardous
// weather warning indicators
type WxButtonStyle struct{}

func (style *WxButtonStyle) Render(view *views.View) WxDataRenderFunction {
	return nil
}

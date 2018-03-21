package main

type TideBar struct{}

// A TideBar displays the current position of the tide, whether it's advancing or
// retreating and its high and low water marks
func (style *TideBar) Render(view *View) WxDataRenderFunction {
	return nil
}

// A WxButton is colored with the current sky color and displays UV and hazardous
// weather warning indicators
type WxButton struct{}

func (style *WxButton) Render(view *View) WxDataRenderFunction {
	return nil
}

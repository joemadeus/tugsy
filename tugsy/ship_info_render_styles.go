package main

type ShipInfo struct{}

func (shipinfo *ShipInfo) Render(view *View) ShipDataRenderFunction {
	return nil
}

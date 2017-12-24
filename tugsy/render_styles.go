package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Render func(history *ShipHistory)

type RenderStyle interface {
	Render(view *View) Render
}

func toRect(position *BaseMapPosition, pixSquare int32) *sdl.Rect {
	return &sdl.Rect{
		X: int32(position.X+0.5) - (pixSquare / 2),
		Y: int32(position.Y+0.5) - (pixSquare / 2),
		W: pixSquare,
		H: pixSquare,
	}
}

func shipTypeToColorMarker(history *ShipHistory) *sdl.Color {
	switch {
	case history.voyagedata == nil:
		positionColor = unknown
	case history.voyagedata.ShipType <= 29:
		positionColor = unknown
	case history.voyagedata.ShipType == 30:
		// fishing
	case history.voyagedata.ShipType <= 32:
		// towing
	case history.voyagedata.ShipType <= 34:
		// diving/dredging/underwater
	case history.voyagedata.ShipType == 35:
		// military ops
	case history.voyagedata.ShipType == 36:
		// sailing
	case history.voyagedata.ShipType == 37:
		// pleasure craft
	case history.voyagedata.ShipType <= 39:
		positionColor = unknown
	case history.voyagedata.ShipType <= 49:
		// high speed craft
	case history.voyagedata.ShipType == 50:
		// pilot vessel
	case history.voyagedata.ShipType == 51:
		// search & rescue
	case history.voyagedata.ShipType == 52:
		// tug
	case history.voyagedata.ShipType == 53:
		// port tender
	case history.voyagedata.ShipType == 54:
		positionColor = unknown // "anti pollution equipment"
	case history.voyagedata.ShipType == 55:
		// law enforcement
	case history.voyagedata.ShipType <= 57:
		positionColor = unknown
	case history.voyagedata.ShipType == 58:
		// medical trnsport
	case history.voyagedata.ShipType == 59:
		// "noncombatant ship"
	case history.voyagedata.ShipType <= 69:
		// passenger
	case history.voyagedata.ShipType <= 79:
		// cargo, dark green
	case history.voyagedata.ShipType <= 89:
		// tanker, light green
	case history.voyagedata.ShipType <= 99:
		positionColor = unknown // other
	}
}

type NullRenderStyle struct{}

func (style *NullRenderStyle) Render(view *View) Render {
	return func(history *ShipHistory) {}
}

type CurrentPositionSimple struct{}

func (style *CurrentPositionSimple) Render(view *View) Render {
	currentPositionColor := sdl.Color{128, 128, 0, 0}
	currentPositionSize := int32(10)
	return func(history *ShipHistory) {
		currentPosition := history.positions[len(history.positions)-1]
		baseMapPosition := view.getBaseMapPosition(currentPosition.GetPositionReport())

		view.screenRenderer.SetDrawColor(currentPositionColor.R, currentPositionColor.G, currentPositionColor.B, sdl.ALPHA_OPAQUE)
		err := view.screenRenderer.FillRect(toRect(&baseMapPosition, currentPositionSize))
		if err != nil {
			logger.Warn("rendering CurrentPositionSimple", "error", err)
		}
	}
}

type CurrentPositionByType struct{}

func (style *CurrentPositionByType) Render(view *View) Render {
	unknown := sdl.Color{128, 128, 128, 0}
	var positionColor sdl.Color
	currentPositionSize := int32(10)

	return func(history *ShipHistory) {

		currentPosition := history.positions[len(history.positions)-1]
		baseMapPosition := view.getBaseMapPosition(currentPosition.GetPositionReport())

		view.screenRenderer.SetDrawColor(positionColor.R, positionColor.G, positionColor.B, sdl.ALPHA_OPAQUE)
		err := view.screenRenderer.FillRect(toRect(&baseMapPosition, currentPositionSize))
		if err != nil {
			logger.Warn("rendering CurrentPositionByType", "error", err)
		}
	}
}

type MarkPathSimple struct{}

func (style *MarkPathSimple) Render(view *View) Render {
	defaultTrackLinesColor := sdl.Color{192, 192, 0, 0}
	defaultTrackPointsColor := sdl.Color{128, 128, 0, 0}
	trackPointsSize := int32(4)

	return func(history *ShipHistory) {
		sdlPoints := make([]sdl.Point, len(history.positions), len(history.positions))
		sdlRects := make([]sdl.Rect, len(history.positions), len(history.positions))

		for i, position := range history.positions {
			baseMapPosition := view.getBaseMapPosition(position.GetPositionReport())
			sdlPoints[i] = sdl.Point{
				X: int32(baseMapPosition.X + 0.5),
				Y: int32(baseMapPosition.Y + 0.5),
			}
			sdlRects[i] = sdl.Rect{
				X: int32(baseMapPosition.X+0.5) - (trackPointsSize / 2),
				Y: int32(baseMapPosition.Y+0.5) - (trackPointsSize / 2),
				W: trackPointsSize,
				H: trackPointsSize,
			}
		}

		view.screenRenderer.SetDrawColor(defaultTrackLinesColor.R, defaultTrackLinesColor.G, defaultTrackLinesColor.B, sdl.ALPHA_OPAQUE)
		err := view.screenRenderer.DrawLines(sdlPoints)
		if err != nil {
			logger.Warn("rendering track lines MarkPathSimple", "error", err)
		}

		view.screenRenderer.SetDrawColor(defaultTrackPointsColor.R, defaultTrackPointsColor.G, defaultTrackPointsColor.B, sdl.ALPHA_OPAQUE)
		err = view.screenRenderer.DrawRects(sdlRects)
		if err != nil {
			logger.Warn("rendering track points MarkPathSimple", "error", err)
		}
	}
}

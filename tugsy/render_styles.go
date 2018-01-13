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

// Returns the SDL color for the given hue (HSV) value, or nil if unknown
func hueToSDLColor(hueVal uint8) *sdl.Color {
	switch hueVal {
	case 0:
		return &sdl.Color{R: 128, G: 128, B: 128, A: 0}
	case 10:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	case 30:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	case 50:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	case 70:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	case 90:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	case 110:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	case 130:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	case 150:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	case 170:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	case 190:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	case 210:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	case 230:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	case 250:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	case 270:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	case 290:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	case 310:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	case 330:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	case 350:
		return &sdl.Color{R: 0, G: 0, B: 0, A: 0}
	default:
		return nil
	}
}

// Maps a ship type to a hue, or to -1 if the type is unknown or it should be
// mapped that way anyway
func shipTypeToHue(history *ShipHistory) uint8 {
	switch {
	case history.voyagedata == nil:
		return -1
	case history.voyagedata.ShipType <= 29:
		return -1
	case history.voyagedata.ShipType == 30:
		// fishing
	case history.voyagedata.ShipType <= 32:
		// towing -- VIOLET: H310
	case history.voyagedata.ShipType <= 34:
		// diving/dredging/underwater
	case history.voyagedata.ShipType == 35:
		// military ops
	case history.voyagedata.ShipType == 36:
		// sailing
	case history.voyagedata.ShipType == 37:
		// pleasure craft
	case history.voyagedata.ShipType <= 39:
		return -1
	case history.voyagedata.ShipType <= 49:
		// high speed craft
	case history.voyagedata.ShipType == 50:
		// pilot vessel -- ORANGE: H50
	case history.voyagedata.ShipType == 51:
		// search & rescue
	case history.voyagedata.ShipType == 52:
		// tug -- RED: H10
	case history.voyagedata.ShipType == 53:
		// port tender -- ORANGE: H50
	case history.voyagedata.ShipType == 54:
		return -1 // "anti pollution equipment"
	case history.voyagedata.ShipType == 55:
		// law enforcement
	case history.voyagedata.ShipType <= 57:
		return -1
	case history.voyagedata.ShipType == 58:
		// medical transport
	case history.voyagedata.ShipType == 59:
		// "noncombatant ship"
	case history.voyagedata.ShipType <= 69:
		// passenger -- GREEN: H110
	case history.voyagedata.ShipType <= 79:
		// cargo -- LIGHT BLUE: H190
	case history.voyagedata.ShipType <= 89:
		// tanker -- DARK BLUE: H250
	case history.voyagedata.ShipType <= 99:
		return -1 // other
	}

	logger.Warn("Mapping an unhandled ship type", "type num", history.voyagedata.ShipType)
	return -1
}

type NullRenderStyle struct{}

func (style *NullRenderStyle) Render(view *View) Render {
	return func(history *ShipHistory) {}
}

type CurrentPositionByType struct{}

func (style *CurrentPositionByType) Render(view *View) Render {
	currentPositionSize := int32(10)
	return func(history *ShipHistory) {

		currentPosition := history.positions[len(history.positions)-1]
		baseMapPosition := view.getBaseMapPosition(currentPosition.GetPositionReport())

		sdlColor := hueToSDLColor(shipTypeToHue(history))
		view.screenRenderer.SetDrawColor(sdlColor.R, sdlColor.G, sdlColor.B, sdl.ALPHA_OPAQUE)
		err := view.screenRenderer.FillRect(toRect(&baseMapPosition, currentPositionSize))
		if err != nil {
			logger.Warn("rendering CurrentPositionByType", "error", err)
		}
	}
}

type MarkPathByType struct{}

func (style *MarkPathByType) Render(view *View) Render {
	trackPointsSize := int32(4)
	trackAlpha := uint8(128)

	return func(history *ShipHistory) {
		pathColor := hueToSDLColor(shipTypeToHue(history))
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

		view.screenRenderer.SetDrawColor(pathColor.R, pathColor.G, pathColor.B, trackAlpha)
		err := view.screenRenderer.DrawLines(sdlPoints)
		if err != nil {
			logger.Warn("rendering track lines MarkPathByType", "error", err)
		}

		err = view.screenRenderer.DrawRects(sdlRects)
		if err != nil {
			logger.Warn("rendering track points MarkPathByType", "error", err)
		}
	}
}

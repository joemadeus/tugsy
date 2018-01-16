package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

const (
	UnknownHue = Hue(361)
	UnknownR   = 128
	UnknownG   = 128
	UnknownB   = 128
)

type Hue uint16
type Render func(history *ShipHistory)

type RenderStyle interface {
	Render(view *View) Render
}

func toDestRect(position *BaseMapPosition, pixSquare int32) *sdl.Rect {
	return &sdl.Rect{
		X: int32(position.X+0.5) - (pixSquare / 2),
		Y: int32(position.Y+0.5) - (pixSquare / 2),
		W: pixSquare,
		H: pixSquare,
	}
}

// Returns the RGB values for the given hue, assuming saturation and value
// equal to 1.0. This is a simplification of the general formula, with C
// equal to 1 and m equal to 0.
//func computeRGB(hue Hue) (r, g, b uint8) {
//
//	// X = C × (1 - |(H / 60°) mod 2 - 1|)
//	x := 1 - math.Abs(hue/60.0 % 2 - 1)
//	X := uint8(x * 255 + 0.5)
//
//	switch {
//	case hue < 60:
//		return 255, X, 0
//	case hue < 120:
//		return X, 255, 0
//	case hue < 180:
//		return 0, 255, X
//	case hue < 240:
//		return 0, X, 255
//	case hue < 300:
//		return X, 0, 255
//	case hue <= 360:
//		return 255, 0, X
//	default:
//		logger.Warn("Got an invalid hue value", "hue", hue)
//		return 128, 128, 128
//	}
//}

// Maps the given Hue value to an RGB triplet, returning neutral gray if
// Hue == 361 (an ordinarily invalid value)
func hueToRGB(hue Hue) (r, g, b uint8) {
	switch hue {
	case 10:
		return 255, 43, 0
	case 30:
		return 255, 128, 0
	case 50:
		return 255, 212, 0
	case 70:
		return 212, 255, 0
	case 90:
		return 128, 255, 0
	case 110:
		return 43, 255, 0
	case 130:
		return 0, 255, 43
	case 150:
		return 0, 255, 128
	case 170:
		return 0, 255, 212
	case 190:
		return 0, 212, 255
	case 210:
		return 0, 128, 255
	case 230:
		return 0, 43, 255
	case 250:
		return 43, 0, 255
	case 270:
		return 128, 0, 255
	case 290:
		return 212, 0, 255
	case 310:
		return 255, 0, 212
	case 330:
		return 255, 0, 128
	case 350:
		return 255, 0, 43
	case UnknownHue:
		return UnknownR, UnknownG, UnknownB
	default:
		logger.Warn("Got an invalid hue value", "hue", hue)
		return UnknownR, UnknownG, UnknownB
	}
}

// Maps a ship type to a hue, or to 0 if the type is unknown or it should be
// mapped that way anyway
func shipTypeToHue(history *ShipHistory) Hue {
	switch {
	case history.voyagedata == nil:
		logger.Debug("voyage data is nil")
		return UnknownHue
	case history.voyagedata.ShipType <= 29:
		return UnknownHue
	case history.voyagedata.ShipType == 30:
		// fishing
	case history.voyagedata.ShipType <= 32:
		// towing -- VIOLET: H310
		return 310
	case history.voyagedata.ShipType <= 34:
		// diving/dredging/underwater
	case history.voyagedata.ShipType == 35:
		// military ops
	case history.voyagedata.ShipType == 36:
		// sailing
	case history.voyagedata.ShipType == 37:
		// pleasure craft -- VIOLET: H290
		return 290
	case history.voyagedata.ShipType <= 39:
		return UnknownHue
	case history.voyagedata.ShipType <= 49:
		// high speed craft -- YELLOW/ORANGE: H50
		return 50
	case history.voyagedata.ShipType == 50:
		// pilot vessel -- ORANGE: H30
		return 30
	case history.voyagedata.ShipType == 51:
		// search & rescue
	case history.voyagedata.ShipType == 52:
		// tug -- RED: H10
		return 10
	case history.voyagedata.ShipType == 53:
		// port tender -- ORANGE: H50
		return 50
	case history.voyagedata.ShipType == 54:
		return UnknownHue // "anti pollution equipment"
	case history.voyagedata.ShipType == 55:
		// law enforcement
	case history.voyagedata.ShipType <= 57:
		return UnknownHue
	case history.voyagedata.ShipType == 58:
		// medical transport
	case history.voyagedata.ShipType == 59:
		// "noncombatant ship"
	case history.voyagedata.ShipType <= 69:
		// passenger -- GREEN: H110
		return 110
	case history.voyagedata.ShipType <= 79:
		// cargo -- LIGHT BLUE: H190
		return 190
	case history.voyagedata.ShipType <= 89:
		// tanker -- DARK BLUE: H250
		return 250
	case history.voyagedata.ShipType <= 99:
		return UnknownHue // other
	}

	logger.Warn("Mapping an unhandled ship type", "type num", history.voyagedata.ShipType)
	return 0
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

		hue := shipTypeToHue(history)
		if hue == UnknownHue && logger.IsDebug() {
			logger.Debug("Ship type is unknown")
		}
		r, g, b := hueToRGB(hue)
		// TODO: Set opacity to 33% if older than a certain age
		view.screenRenderer.SetDrawColor(r, g, b, sdl.ALPHA_OPAQUE)

		err := view.screenRenderer.FillRect(toDestRect(&baseMapPosition, currentPositionSize))
		if err != nil {
			logger.Warn("rendering CurrentPositionByType", "error", err)
		}
	}
}

type CurrentPositionByTypeSprite struct {
	SpecialSprites *Special
	DotSprites     *Dots
}

func NewCurrentPositionByTypeSprite(screenRenderer *sdl.Renderer) (*CurrentPositionByTypeSprite, error) {
	special, err := NewSpecial(screenRenderer)
	if err != nil {
		return nil, err
	}

	dots, err := NewDots(screenRenderer)
	if err != nil {
		return nil, err
	}

	return &CurrentPositionByTypeSprite{SpecialSprites: special, DotSprites: dots}, nil
}

func (style *CurrentPositionByTypeSprite) Render(view *View) Render {
	return func(history *ShipHistory) {

		currentPosition := history.positions[len(history.positions)-1]
		baseMapPosition := view.getBaseMapPosition(currentPosition.GetPositionReport())

		hue := shipTypeToHue(history)
		var srcRect *sdl.Rect
		var sheet *SpriteSheet
		var ok bool
		if hue == UnknownHue {
			srcRect, sheet, ok = style.SpecialSprites.GetSprite("unknown")
			if ok == false {
				return
			}
		} else {
			srcRect, sheet, ok = style.DotSprites.GetSprite(hue, "normal")
			if ok == false {
				return
			}
		}

		// TODO: Set opacity to 33% if older than a certain age

		err := view.screenRenderer.Copy(
			sheet.Texture, srcRect, toDestRect(&baseMapPosition, sheet.SpriteSize))
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
		r, g, b := hueToRGB(shipTypeToHue(history))
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

		view.screenRenderer.SetDrawColor(r, g, b, trackAlpha)
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

package shipdata

import (
	"github.com/joemadeus/tugsy/tugsy/config"
	"github.com/joemadeus/tugsy/tugsy/views"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	defaultDestSpriteSizePixels = 20
)

type ShipDataRenderFunction func(history *ShipHistory)

func toDestRect(position *views.BaseMapPosition, pixSquare int32) *sdl.Rect {
	return &sdl.Rect{
		X: int32(position.X+0.5) - (pixSquare / 2),
		Y: int32(position.Y+0.5) - (pixSquare / 2),
		W: pixSquare,
		H: pixSquare,
	}
}

// Maps a ship type to a hue, or to UnknownHue if the type is unknown or should be
// mapped that way anyway
func shipTypeToHue(history *ShipHistory) views.Hue {
	switch {
	case history.voyagedata == nil:
		return views.UnknownHue
	case history.voyagedata.ShipType <= 29:
		return views.UnknownHue
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
		return views.UnknownHue
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
		return views.UnknownHue // "anti pollution equipment"
	case history.voyagedata.ShipType == 55:
		// law enforcement
	case history.voyagedata.ShipType <= 57:
		return views.UnknownHue
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
		return views.UnknownHue // other
	}

	logger.Warn("Mapping an unhandled ship type", "type num", history.voyagedata.ShipType)
	return 0
}

type NullRenderStyle struct{}

func (style *NullRenderStyle) Render(view *views.View) func() {
	return func() {}
}

type CurrentPositionByTypeStyle struct {
	SpecialSprites *views.Special
	DotSprites     *views.Dots
}

func NewCurrentPositionByType(screenRenderer *sdl.Renderer, config *config.Config) (*CurrentPositionByTypeStyle, error) {
	special, err := views.NewSpecial(screenRenderer, config)
	if err != nil {
		return nil, err
	}

	dots, err := views.NewDots(screenRenderer, config)
	if err != nil {
		return nil, err
	}

	return &CurrentPositionByTypeStyle{SpecialSprites: special, DotSprites: dots}, nil
}

func (style *CurrentPositionByTypeStyle) Render(view *views.View) ShipDataRenderFunction {
	return func(history *ShipHistory) {

		currentPosition := history.positions[len(history.positions)-1]
		baseMapPosition := view.GetBaseMapPosition(currentPosition.GetPositionReport())

		hue := shipTypeToHue(history)
		var srcRect *sdl.Rect
		var sheet *views.SpriteSheet
		var ok bool
		if hue == views.UnknownHue {
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

		err := view.ScreenRenderer.Copy(
			sheet.Texture, srcRect, toDestRect(&baseMapPosition, defaultDestSpriteSizePixels))
		if err != nil {
			logger.Warn("rendering CurrentPositionByType", "error", err)
		}
	}
}

type MarkPathByType struct{}

func NewMarkPathByType() (*MarkPathByType, error) {
	return &MarkPathByType{}, nil
}

func (style *MarkPathByType) Render(view *views.View) ShipDataRenderFunction {
	trackPointsSize := int32(4)
	trackAlpha := uint8(128)

	return func(history *ShipHistory) {
		r, g, b := views.HueToRGB(shipTypeToHue(history))
		sdlPoints := make([]sdl.Point, len(history.positions), len(history.positions))
		sdlRects := make([]sdl.Rect, len(history.positions), len(history.positions))

		for i, position := range history.positions {
			baseMapPosition := view.GetBaseMapPosition(position.GetPositionReport())
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

		view.ScreenRenderer.SetDrawColor(r, g, b, trackAlpha)
		err := view.ScreenRenderer.DrawLines(sdlPoints)
		if err != nil {
			logger.Warn("rendering track lines MarkPathByType", "error", err)
		}

		err = view.ScreenRenderer.DrawRects(sdlRects)
		if err != nil {
			logger.Warn("rendering track points MarkPathByType", "error", err)
		}
	}
}

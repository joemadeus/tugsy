package shipdata

import (
	"github.com/joemadeus/tugsy/tugsy/views"
	logger "github.com/sirupsen/logrus"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	defaultDestSpriteSizePixels = 20
)

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

	logger.WithField("type num", history.voyagedata.ShipType).Warn("mapping an unhandled ship type")
	return 0
}

type ShipPositionElement struct {
	aisData        *AISData
	SpecialSprites *views.SpecialSheet
	DotSprites     *views.DotSheet
}

func NewShipPositionElement(ais *AISData, spriteSet *views.SpriteSet) *ShipPositionElement {
	logger.Info("Loading 'Ship Position' element")
	return &ShipPositionElement{aisData: ais, SpecialSprites: spriteSet.SpecialSheet, DotSprites: spriteSet.DotSheet}
}

// ShipPositionElement renders ship tracks and positions
func (ele *ShipPositionElement) Render(view *views.View) error {
	for _, mmsi := range ele.aisData.GetHistoryMMSIs() {
		// lock held in GetShipHistory
		history, ok := ele.aisData.GetShipHistory(mmsi)

		if ok == false {
			logger.Debug("history has vanished", "mmsi", mmsi)
			continue
		}

		history.Lock()
		if len(history.positions) == 0 {
			logger.Debug("no positions", "mmsi", mmsi)
			continue
		}

		ele.renderPath(view, history)
		ele.renderCurrentPosition(view, history)

		history.dirty = false
		history.Unlock()
	}

	return nil
}

func (ele *ShipPositionElement) renderPath(view *views.View, history *ShipHistory) error {
	currentPosition := history.positions[len(history.positions)-1]
	baseMapPosition := view.GetBaseMapPosition(currentPosition.GetPositionReport())

	hue := shipTypeToHue(history)
	var sprite *views.Sprite
	var err error
	if hue == views.UnknownHue {
		// return the special "unknown" dot
		if sprite, err = ele.SpecialSprites.GetSprite("unknown"); err != nil {
			logger.WithError(err).Error("could not load special sprite 'unknown'")
			return err
		}
	} else {
		if sprite, err = ele.DotSprites.GetSprite(hue, "normal"); err != nil {
			logger.WithError(err).Errorf("could not load sprite with hue %v", hue)
			return err
		}
	}

	// TODO: Set opacity to 33% if older than a certain age

	// TODO: Set hazardous cargo markers

	if err := view.ScreenRenderer.Copy(sprite.Texture, sprite.Rect, toDestRect(&baseMapPosition, defaultDestSpriteSizePixels)); err != nil {
		logger.WithError(err).Error("rendering ship paths")
		return err
	}

	return nil
}

func (ele *ShipPositionElement) renderCurrentPosition(view *views.View, history *ShipHistory) error {
	trackPointsSize := int32(4)
	trackAlpha := uint8(128)

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

	if err := view.ScreenRenderer.SetDrawColor(r, g, b, trackAlpha); err != nil {
		logger.WithError(err).Warn("setting the draw color")
		return err
	}

	if err := view.ScreenRenderer.DrawLines(sdlPoints); err != nil {
		logger.WithError(err).Warn("rendering track lines")
		return err
	}

	if err := view.ScreenRenderer.DrawRects(sdlRects); err != nil {
		logger.WithError(err).Warn("rendering track points")
		return err
	}

	return nil
}

package shipdata

import (
	"math"
	"sync"

	"github.com/joemadeus/tugsy/tugsy/views"
	"github.com/sirupsen/logrus"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	defaultDestSpriteSizePixels = 20
)

type screenPosition struct {
	views.BaseMapPosition
	mmsi uint32
}

type ShipPositionElement struct {
	sync.Mutex
	aisData        *AISData
	curPositions   []screenPosition
	specialSprites *views.SpecialSheet
	dotSprites     *views.DotSheet
}

func NewShipPositionElement(ais *AISData, spriteSet *views.SpriteSet) *ShipPositionElement {
	logrus.Info("Loading 'Ship Position' element")
	return &ShipPositionElement{aisData: ais, specialSprites: spriteSet.SpecialSheet, dotSprites: spriteSet.DotSheet}
}

// ScreenLookup finds the MMSI for the ship closest to the provided X and Y screen
// coordinates. Returns zero if there are no ships within the 'within' distance
func (e *ShipPositionElement) Within(x, y int32, fluff float64) bool {
	// for now we're just doing this linearly. we handle these sort of lookups only
	// infrequently and the cardinality isn't usually high
	closest := struct {
		mmsi uint32
		d    float64
	}{d: math.MaxFloat64}
	for _, sp := range e.curPositions {
		distance := math.Sqrt(math.Pow(float64(x)-sp.X, 2) + math.Pow(float64(y)-sp.Y, 2))
		if distance > fluff {
			continue
		}

		if distance > closest.d {
			continue
		}

		if _, there := e.aisData.GetShipHistory(sp.mmsi); there == false {
			continue
		}
	}

	return closest.mmsi
}

// Render renders ship tracks and positions
func (e *ShipPositionElement) Render(view *views.View) error {
	rendered := make([]screenPosition, 0)
	for _, sh := range e.aisData.GetHistoryMMSIs() {
		if len(sh.positions) == 0 {
			logrus.Debug("no positions", "mmsi", sh.mmsi)
			continue
		}

		if err := e.renderHistory(view, sh); err != nil {
			return err
		}

		p, err := e.renderPosition(view, sh)
		if err != nil {
			return err
		}
		rendered = append(rendered, p)
	}

	e.Lock()
	defer e.Unlock()
	e.curPositions = rendered

	return nil
}

func (e *ShipPositionElement) renderPosition(view *views.View, history *ShipHistory) (screenPosition, error) {
	currentPosition := history.positions[len(history.positions)-1]
	baseMapPosition := view.GetBaseMapPosition(currentPosition.GetPositionReport())

	hue := shipTypeToHue(history)
	var sprite *views.Sprite
	var err error
	if hue == views.UnknownHue {
		// return the special "unknown" dot
		if sprite, err = e.specialSprites.GetSprite("unknown"); err != nil {
			logrus.WithError(err).Error("could not load special sprite 'unknown'")
			return screenPosition{}, err
		}
	} else {
		if sprite, err = e.dotSprites.GetSprite(hue, "normal"); err != nil {
			logrus.WithError(err).Errorf("could not load sprite with hue %v", hue)
			return screenPosition{}, err
		}
	}

	// TODO: Set opacity to 33% if older than a certain age

	// TODO: Set hazardous cargo markers

	if err := view.ScreenRenderer.Copy(sprite.Texture, sprite.Rect, toDestRect(&baseMapPosition, defaultDestSpriteSizePixels)); err != nil {
		logrus.WithError(err).Error("rendering ship history")
		return screenPosition{}, err
	}

	return screenPosition{BaseMapPosition: baseMapPosition, mmsi: history.mmsi}, nil
}

func (e *ShipPositionElement) renderHistory(view *views.View, history *ShipHistory) error {
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
		logrus.WithError(err).Warn("setting the draw color")
		return err
	}

	if err := view.ScreenRenderer.DrawLines(sdlPoints); err != nil {
		logrus.WithError(err).Warn("rendering track lines")
		return err
	}

	if err := view.ScreenRenderer.DrawRects(sdlRects); err != nil {
		logrus.WithError(err).Warn("rendering track points")
		return err
	}

	return nil
}

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

	logrus.WithField("type num", history.voyagedata.ShipType).Warn("mapping an unhandled ship type")
	return 0
}

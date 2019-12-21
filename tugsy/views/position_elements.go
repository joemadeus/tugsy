package views

import (
	"math"
	"sync"

	"github.com/joemadeus/tugsy/tugsy/shipdata"
	"github.com/sirupsen/logrus"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	defaultDestSpriteSizePixels = 20
)

type screenPosition struct {
	BaseMapPosition
	mmsi uint32
}

type ShipPositionElement struct {
	sync.Mutex
	aisData        *shipdata.AISData
	curPositions   []screenPosition
	specialSprites *SpecialSheet
	dotSprites     *DotSheet
}

func NewShipPositionElement(ais *shipdata.AISData, spriteSet *SpriteSet) *ShipPositionElement {
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

		if _, there := e.aisData.ShipHistory(sp.mmsi); there == false {
			continue
		}
	}

	return closest.mmsi
}

// Render renders ship tracks and positions
func (e *ShipPositionElement) Render(view *View) error {
	rendered := make([]screenPosition, 0)
	for _, sh := range e.aisData.Historys() {
		positions := sh.Positions()
		hue := shipTypeToHue(sh)
		if len(positions) == 0 {
			logrus.Debug("no positions", "mmsi", sh.MMSI)
			continue
		}

		if err := e.renderHistory(view, hue, positions); err != nil {
			return err
		}

		p, err := e.renderPosition(view, hue, positions)
		if err != nil {
			return err
		}

		rendered = append(rendered, screenPosition{p, sh.MMSI})
	}

	e.Lock()
	defer e.Unlock()
	e.curPositions = rendered

	return nil
}

func (e *ShipPositionElement) renderPosition(view *View, hue Hue, positions []shipdata.Positionable) (BaseMapPosition, error) {
	currentPosition := positions[0]
	baseMapPosition := view.GetBaseMapPosition(currentPosition.GetPositionReport())

	var sprite *Sprite
	var err error
	if hue == UnknownHue {
		// return the special "unknown" dot
		if sprite, err = e.specialSprites.GetSprite("unknown"); err != nil {
			logrus.WithError(err).Error("could not load special sprite 'unknown'")
			return BaseMapPosition{}, err
		}
	} else {
		if sprite, err = e.dotSprites.GetSprite(hue, "normal"); err != nil {
			logrus.WithError(err).Errorf("could not load sprite with hue %v", hue)
			return BaseMapPosition{}, err
		}
	}

	// TODO: Set opacity to 33% if older than a certain age

	// TODO: Set hazardous cargo markers

	if err := view.ScreenRenderer.Copy(sprite.Texture, sprite.Rect, toDestRect(&baseMapPosition, defaultDestSpriteSizePixels)); err != nil {
		logrus.WithError(err).Error("rendering ship history")
		return BaseMapPosition{}, err
	}

	return baseMapPosition, nil
}

func (e *ShipPositionElement) renderHistory(view *View, hue Hue, positions []shipdata.Positionable) error {
	trackPointsSize := int32(4)
	trackAlpha := uint8(128)

	r, g, b := HueToRGB(hue)
	sdlPoints := make([]sdl.Point, len(positions), len(positions))
	sdlRects := make([]sdl.Rect, len(positions), len(positions))

	for i, position := range positions {
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

func toDestRect(position *BaseMapPosition, pixSquare int32) *sdl.Rect {
	return &sdl.Rect{
		X: int32(position.X+0.5) - (pixSquare / 2),
		Y: int32(position.Y+0.5) - (pixSquare / 2),
		W: pixSquare,
		H: pixSquare,
	}
}

// Maps a ship type to a hue, or to UnknownHue if the type is unknown or should be
// mapped that way anyway
func shipTypeToHue(history *shipdata.ShipHistory) Hue {
	voyagedata := history.VoyageData()
	switch {
	case voyagedata == nil:
		return UnknownHue
	case voyagedata.ShipType <= 29:
		return UnknownHue
	case voyagedata.ShipType == 30:
		// fishing
	case voyagedata.ShipType <= 32:
		// towing -- VIOLET: H310
		return 310
	case voyagedata.ShipType <= 34:
		// diving/dredging/underwater
	case voyagedata.ShipType == 35:
		// military ops
	case voyagedata.ShipType == 36:
		// sailing
	case voyagedata.ShipType == 37:
		// pleasure craft -- VIOLET: H290
		return 290
	case voyagedata.ShipType <= 39:
		return UnknownHue
	case voyagedata.ShipType <= 49:
		// high speed craft -- YELLOW/ORANGE: H50
		return 50
	case voyagedata.ShipType == 50:
		// pilot vessel -- ORANGE: H30
		return 30
	case voyagedata.ShipType == 51:
		// search & rescue
	case voyagedata.ShipType == 52:
		// tug -- RED: H10
		return 10
	case voyagedata.ShipType == 53:
		// port tender -- ORANGE: H50
		return 50
	case voyagedata.ShipType == 54:
		return UnknownHue // "anti pollution equipment"
	case voyagedata.ShipType == 55:
		// law enforcement
	case voyagedata.ShipType <= 57:
		return UnknownHue
	case voyagedata.ShipType == 58:
		// medical transport
	case voyagedata.ShipType == 59:
		// "noncombatant ship"
	case voyagedata.ShipType <= 69:
		// passenger -- GREEN: H110
		return 110
	case voyagedata.ShipType <= 79:
		// cargo -- LIGHT BLUE: H190
		return 190
	case voyagedata.ShipType <= 89:
		// tanker -- DARK BLUE: H250
		return 250
	case voyagedata.ShipType <= 99:
		return UnknownHue // other
	}

	logrus.WithField("type num", voyagedata.ShipType).Warn("mapping an unhandled ship type")
	return 0
}

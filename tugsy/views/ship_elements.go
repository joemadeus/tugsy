package views

import (
	"math"
	"reflect"
	"sync"

	"github.com/joemadeus/tugsy/tugsy/shipdata"
	logger "github.com/sirupsen/logrus"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	defaultDestSpriteSizePixels = 20
)

// ShipInfoElement renders information for a ship, including its registration, flag,
// name, current destination and current situation (moored, underway, etc) into a
// BaseInfoElement
type ShipInfoElement struct {
	flag    ChildElement
	history *shipdata.ShipHistory
}

func NewShipInfoElement(h *shipdata.ShipHistory) *ShipInfoElement {
	return &ShipInfoElement{history: h}
}

func (e *ShipInfoElement) ClosestChild(x, y int32) (ChildElement, float64) {
	return e.flag, e.flag.Distance(x, y)
}

func (e *ShipInfoElement) Render(v *View) error {
	// TODO
	return nil
}

type AllPositionElements struct {
	sync.Mutex
	*SpriteSet

	aisData          *shipdata.AISData
	positionElements map[uint32]*ShipPositionElement
	baseInfoElement  *BaseInfoElement
}

func NewAllPositionElements(sprites *SpriteSet, ais *shipdata.AISData, be *BaseInfoElement) *AllPositionElements {
	return &AllPositionElements{
		SpriteSet:        sprites,
		aisData:          ais,
		positionElements: make(map[uint32]*ShipPositionElement),
		baseInfoElement:  be,
	}
}

func (e *AllPositionElements) ClosestChild(x, y int32) (ChildElement, float64) {
	e.Lock()
	defer e.Unlock()

	closest := struct {
		ele *ShipPositionElement
		d   float64
	}{d: math.MaxFloat64}
	for _, sp := range e.positionElements {
		d := sp.Distance(x, y)
		if d > closest.d {
			continue
		}

		closest.d = d
		closest.ele = sp
	}

	logger.Debugf("AllPositionElements ClosestChild at %s, %f", reflect.TypeOf(closest.ele).String(), closest.d)
	return closest.ele, closest.d
}

func (e *AllPositionElements) Render(v *View) error {
	mmsis := make(map[uint32]struct{})
	histories := e.aisData.ShipHistories() // returns a copy

	e.Lock()
	defer e.Unlock()

	for _, sh := range histories {
		se, ok := e.positionElements[sh.MMSI]
		if ok == false {
			se = &ShipPositionElement{SpriteSet: e.SpriteSet, history: sh, baseInfoElement: e.baseInfoElement}
			e.positionElements[sh.MMSI] = se
		}

		if err := se.Render(v); err != nil {
			return err
		}

		mmsis[sh.MMSI] = struct{}{}
	}

	// prune ShipPositionElements that no longer exist. could be replaced, along
	// with add(), with a chan, I suppose
	for m := range e.positionElements {
		if _, ok := mmsis[m]; ok == false {
			delete(e.positionElements, m)
		}
	}

	return nil
}

type ShipPositionElement struct {
	*SpriteSet

	curPosition     BaseMapPosition
	history         *shipdata.ShipHistory
	baseInfoElement *BaseInfoElement
}

func (e *ShipPositionElement) Distance(x, y int32) float64 {
	d := screenDistance(x, y, e.curPosition.X, e.curPosition.Y)
	logger.Debugf("ShipPositionElement distance is %f", d)
	return d
}

func (e *ShipPositionElement) HandleTouch() error {
	logger.Debug("Handling touch in ShipPositionElement")
	infoElement := NewShipInfoElement(e.history)
	return e.baseInfoElement.UpdateContent(infoElement)
}

func (e *ShipPositionElement) Render(v *View) error {
	positions := e.history.Positions()
	hue := shipTypeToHue(e.history)
	if len(positions) == 0 {
		return nil
	}

	// TODO we're reloading sprites and primitives every time through. cut that
	//  out and start holding some view state

	if err := e.renderHistory(v, hue, positions); err != nil {
		return err
	}

	if err := e.renderPosition(v, hue, positions); err != nil {
		return err
	}

	return nil
}

func (e *ShipPositionElement) renderPosition(view *View, hue Hue, positions []shipdata.Positionable) error {
	e.curPosition = view.BaseMapPosition(positions[len(positions)-1].GetPositionReport())

	var sprite *Sprite
	var err error
	if hue == UnknownHue {
		// return the special "unknown" dot
		if sprite, err = e.SpecialSheet.GetSprite("unknown"); err != nil {
			logger.WithError(err).Error("could not load special sprite 'unknown'")
			return err
		}
	} else {
		if sprite, err = e.DotSheet.GetSprite(hue, "normal"); err != nil {
			logger.WithError(err).Errorf("could not load sprite with hue %v", hue)
			return err
		}
	}

	// TODO: Set opacity to 33% if older than a certain age

	// TODO: Set hazardous cargo markers

	if err := view.ScreenRenderer.Copy(sprite.Texture, sprite.Rect, toDestRect(&e.curPosition, defaultDestSpriteSizePixels)); err != nil {
		logger.WithError(err).Error("rendering ship history")
		return err
	}

	return nil
}

func (e *ShipPositionElement) renderHistory(view *View, hue Hue, positions []shipdata.Positionable) error {
	trackPointsSize := int32(4)
	trackAlpha := uint8(128)

	r, g, b := HueToRGB(hue)
	sdlPoints := make([]sdl.Point, len(positions), len(positions))
	sdlRects := make([]sdl.Rect, len(positions), len(positions))

	for i, position := range positions {
		baseMapPosition := view.BaseMapPosition(position.GetPositionReport())
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

	logger.WithField("type num", voyagedata.ShipType).Warn("mapping an unhandled ship type")
	return 0
}

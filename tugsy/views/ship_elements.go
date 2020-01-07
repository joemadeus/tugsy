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

// ShipInfoElement renders information for a ship, including its registration, flag,
// name, current destination and current situation (moored, underway, etc) into a
// BaseInfoElement
type ShipInfoElement struct {
	flag        ChildElement
	port        ChildElement
	infoElement *BaseInfoElement
	history     *shipdata.ShipHistory
}

func (e *ShipInfoElement) ClosestChild(x, y int32) (ChildElement, float64) {
	fD := e.flag.Distance(x, y)
	pD := e.port.Distance(x, y)
	if fD < pD {
		return e.flag, fD
	}

	return e.port, pD
}

func (e *ShipInfoElement) Render(v *View) error {
	panic("implement me")
}

func (e *ShipInfoElement) Update(h *shipdata.ShipHistory) error {
	return e.infoElement.UpdateContent(e)
}

type AllPositionElements struct {
	sync.Mutex
	*SpriteSet

	aisData          *shipdata.AISData
	positionElements map[uint32]*ShipPositionElement
}

func NewAllPositionElements(sprites *SpriteSet, ais *shipdata.AISData) *AllPositionElements {
	return &AllPositionElements{
		SpriteSet:        sprites,
		aisData:          ais,
		positionElements: make(map[uint32]*ShipPositionElement),
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
			se = &ShipPositionElement{SpriteSet: e.SpriteSet, history: sh}
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

func (e *AllPositionElements) SetHistory(h *shipdata.ShipHistory) {
	e.Lock()
	defer e.Unlock()
	e.positionElements[h.MMSI] = &ShipPositionElement{
		history:   h,
		SpriteSet: e.SpriteSet,
	}
}

func (e *AllPositionElements) RemoveHistory(h *shipdata.ShipHistory) {
	e.Lock()
	defer e.Unlock()
	delete(e.positionElements, h.MMSI)
}

type ShipPositionElement struct {
	*SpriteSet

	history     *shipdata.ShipHistory
	curPosition BaseMapPosition
	infoElement *ShipInfoElement
}

func (e *ShipPositionElement) Distance(x, y int32) float64 {
	// this is a leaf -- there are no children
	return screenDistance(x, y, e.curPosition.X, e.curPosition.Y)
}

func (e *ShipPositionElement) HandleTouch() error {
	e.infoElement.history = e.history
	return nil
}

func (e *ShipPositionElement) Render(v *View) error {
	positions := e.history.Positions()
	hue := shipTypeToHue(e.history)
	if len(positions) == 0 {
		logrus.Debug("no positions", "mmsi", e.history.MMSI)
		return nil
	}

	if err := e.renderHistory(v, hue, positions); err != nil {
		return err
	}

	p, err := e.renderPosition(v, hue, positions)
	if err != nil {
		return err
	}

	e.curPosition = p
	return nil
}

func (e *ShipPositionElement) renderPosition(view *View, hue Hue, positions []shipdata.Positionable) (BaseMapPosition, error) {
	currentPosition := positions[len(positions)-1]
	baseMapPosition := view.BaseMapPosition(currentPosition.GetPositionReport())

	var sprite *Sprite
	var err error
	if hue == UnknownHue {
		// return the special "unknown" dot
		if sprite, err = e.SpecialSheet.GetSprite("unknown"); err != nil {
			logrus.WithError(err).Error("could not load special sprite 'unknown'")
			return BaseMapPosition{}, err
		}
	} else {
		if sprite, err = e.DotSheet.GetSprite(hue, "normal"); err != nil {
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

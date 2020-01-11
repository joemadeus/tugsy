package views

import (
	"math"
	"reflect"
	"sync"

	"github.com/joemadeus/tugsy/tugsy/config"
	logger "github.com/sirupsen/logrus"
	image "github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	infoBackgroundFile         = "infoBackground.png"
	infoBorderFile             = "infoBorder.png"
	infoPaneDstX, infoPaneDstY = 230, 10
	infoPaneSrcX, infoPaneSrcY = 0, 0
	infoPaneH, infoPaneW       = 120, 240

	closeButtonFile                  = "close.png"
	closeButtonDstX, closeButtonDstY = 205, 35
	closeButtonSrcX, closeButtonSrcY = 0, 0
	closeButtonH, closeButtonW       = 20, 20
)

type UIElement interface {
	// Render causes whatever graphics are appropriate for the element to be rendered
	// to the screen
	Render(*View) error
}

// ParentElements are UI elements that contain other UI elements. They are not, by
// themselves, able to respond to user input, but given an X and Y will recursively
// iterate through their member elements and return the ChildElement that is able
// to respond (if any)
type ParentElement interface {
	UIElement

	// ClosestChild returns the element closest (on screen) to the given X & Y,
	// with its distance
	ClosestChild(x, y int32) (ChildElement, float64)
}

// ChildElements are UI elements that can respond to interaction with the user
type ChildElement interface {
	UIElement

	// HandleTouch performs whatever action is needed when the user touches or selects
	// the element on the screen
	HandleTouch() error

	// Distance returns the minimal distance between this child and the given
	// X & Y screen coordinates
	Distance(x, y int32) float64
}

type RootElement struct {
	baseInfoElement     ParentElement
	allPositionsElement ParentElement
	// wxElement           *WxElement

	touchFluff float64
}

func NewRootElement(cfg *config.Config, baseEle, allPosEle ParentElement) *RootElement {
	return &RootElement{
		touchFluff:          cfg.GetFloat64("touchFluff"),
		baseInfoElement:     baseEle,
		allPositionsElement: allPosEle,
	}
}

func (e *RootElement) ClosestChild(x, y int32) (ChildElement, float64) {
	closest := struct {
		ele ChildElement
		d   float64
	}{d: math.MaxFloat64}

	for _, ele := range []ParentElement{e.baseInfoElement, e.allPositionsElement} {
		e, d := ele.ClosestChild(x, y)
		if d > closest.d {
			continue
		}

		closest.ele = e
		closest.d = d
	}

	return closest.ele, closest.d
}

func (e *RootElement) Render(v *View) error {
	for _, ele := range []UIElement{e.baseInfoElement, e.allPositionsElement} {
		if err := ele.Render(v); err != nil {
			return err
		}
	}

	return nil
}

// Touch finds the ChildElement whose Distance() is closest to the given X & Y and
// is within the RootElement's touchFluff value. If one is found, the ChildELement's
// HandleTouch() is invoked
func (e *RootElement) Touch(x, y int32) error {
	logger.Debugf("RootElement touch at %d, %d", x, y)
	closest, dist := e.ClosestChild(x, y)
	if dist > e.touchFluff {
		return nil
	}

	return closest.HandleTouch()
}

// A BaseInfoElement is the "info pane" in the upper right part of the UI. It has a
// two child UIElements: 'content', which displays the actual info (this type is
// simply the frame in which that content is displayed) and 'close', which removes
// the content and disables the info pane's rendering.
type BaseInfoElement struct {
	sync.Mutex

	content ParentElement
	close   ChildElement

	Background  *sdl.Texture
	Border      *sdl.Texture
	BaseSrcRect *sdl.Rect
	BaseDstRect *sdl.Rect
}

func NewBaseInfoElement(cfg *config.Config, renderer *sdl.Renderer) (*BaseInfoElement, error) {
	logger.Info("Loading 'Info' element")
	ele := &BaseInfoElement{
		BaseSrcRect: &sdl.Rect{X: infoPaneSrcX, Y: infoPaneSrcY, H: infoPaneH, W: infoPaneW},
		BaseDstRect: &sdl.Rect{X: infoPaneDstX, Y: infoPaneDstY, H: infoPaneH, W: infoPaneW},
	}

	var err error
	ele.Background, err = image.LoadTexture(renderer, cfg.SpriteSheetPath(infoBackgroundFile))
	if err != nil {
		return nil, err
	}

	ele.Border, err = image.LoadTexture(renderer, cfg.SpriteSheetPath(infoBorderFile))
	if err != nil {
		return nil, err
	}

	ele.close, err = NewCloseElement(cfg, renderer, ele)
	if err != nil {
		return nil, err
	}

	return ele, nil
}

func (e *BaseInfoElement) ClosestChild(x, y int32) (ChildElement, float64) {
	e.Lock()
	defer e.Unlock()

	if e.content == nil {
		return nil, math.MaxFloat64
	}

	// TODO: shortcut... test that the point specified is within the BaseInfoElement's
	//  rect before traversing its tree of children

	contentChild, contentD := e.content.ClosestChild(x, y)
	closeD := e.close.Distance(x, y)
	if closeD < contentD {
		logger.Debugf("BaseInfoElement ClosestChild at %s, %f", "CloseElement", closeD)
		return e.close, closeD
	}

	logger.Debugf("BaseInfoElement ClosestChild at %s, %f", reflect.TypeOf(contentChild).String(), contentD)
	return contentChild, contentD
}

func (e *BaseInfoElement) Render(v *View) error {
	e.Lock()
	defer e.Unlock()

	// don't bother showing the info element if there's no content
	if e.content == nil {
		return nil
	}

	if err := v.ScreenRenderer.Copy(e.Background, e.BaseSrcRect, e.BaseDstRect); err != nil {
		return err
	}

	if err := e.content.Render(v); err != nil {
		return err
	}

	if err := v.ScreenRenderer.Copy(e.Border, e.BaseSrcRect, e.BaseDstRect); err != nil {
		return err
	}

	return nil
}

func (e *BaseInfoElement) UpdateContent(c ParentElement) error {
	e.Lock()
	defer e.Unlock()

	if c == nil {
		logger.Debug("clearing info pane")
	} else {
		logger.Debug("setting info pane")
	}

	e.content = c
	return nil
}

// A CloseElement responds to user input by disabling info pane rendering
type CloseElement struct {
	closedElement *BaseInfoElement

	tex        *sdl.Texture
	texSrcRect *sdl.Rect
	texDstRect *sdl.Rect
}

func NewCloseElement(cfg *config.Config, renderer *sdl.Renderer, base *BaseInfoElement) (*CloseElement, error) {
	cl := &CloseElement{
		closedElement: base,
		texSrcRect:    &sdl.Rect{X: closeButtonSrcX, Y: closeButtonSrcY, H: closeButtonH, W: closeButtonW},
		texDstRect:    &sdl.Rect{X: closeButtonDstX, Y: closeButtonDstY, H: closeButtonH, W: closeButtonW},
	}

	var err error
	cl.tex, err = image.LoadTexture(renderer, cfg.SpriteSheetPath(closeButtonFile))
	if err != nil {
		return nil, err
	}

	return cl, nil
}

// Distance calculates the screen distance from the center of the close icon's dst rect
func (e *CloseElement) Distance(x, y int32) float64 {
	d := screenDistance(x, y, closeButtonW/2.0+closeButtonDstX, closeButtonH/2.0+closeButtonDstY)
	logger.Debugf("CloseElement distance is %f", d)
	return d
}

func (e *CloseElement) HandleTouch() error {
	return e.closedElement.UpdateContent(nil)
}

func (e *CloseElement) Render(v *View) error {
	return v.ScreenRenderer.Copy(e.tex, e.texSrcRect, e.texDstRect)
}

func screenDistance(sX, sY int32, tX, tY float64) float64 {
	return math.Sqrt(math.Pow(float64(sX)-tX, 2) + math.Pow(float64(sY)-tY, 2))
}

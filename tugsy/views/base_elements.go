package views

import (
	"sync"

	"github.com/joemadeus/tugsy/tugsy/config"
	logger "github.com/sirupsen/logrus"
	image "github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	infoBackgroundFile = "infoBackground.png"
	infoBorderFile     = "infoBorder.png"

	infoPaneDstX = 230
	infoPaneDstY = 10
	infoPaneSrcX = 0
	infoPaneSrcY = 0
	infoPaneH    = 120
	infoPaneW    = 240
)

// A InfoElement has a base graphic with static size and position onto/into which
// other graphics are overlaid
type InfoElement struct {
	sync.Mutex

	Content    ViewElement
	Background *sdl.Texture
	Border     *sdl.Texture

	BaseSrcRect *sdl.Rect
	BaseDstRect *sdl.Rect
}

func NewInfoElement(screenRenderer *sdl.Renderer, config *config.Config) (*InfoElement, error) {
	logger.Info("Loading 'Info' element")
	backgroundTex, err := image.LoadTexture(screenRenderer, config.GetSpritesheetPath(infoBackgroundFile))
	if err != nil {
		return nil, err
	}

	borderTex, err := image.LoadTexture(screenRenderer, config.GetSpritesheetPath(infoBorderFile))
	if err != nil {
		return nil, err
	}

	return &InfoElement{
		Background: backgroundTex,
		Border:     borderTex,

		BaseSrcRect: &sdl.Rect{X: infoPaneSrcX, Y: infoPaneSrcY, H: infoPaneH, W: infoPaneW},
		BaseDstRect: &sdl.Rect{X: infoPaneDstX, Y: infoPaneDstY, H: infoPaneH, W: infoPaneW},
	}, nil
}

func (ie *InfoElement) Render(view *View) error {
	ie.Lock()
	defer ie.Unlock()

	// don't bother showing the info element if there's no content
	if ie.Content == nil {
		return nil
	}

	if err := view.ScreenRenderer.Copy(ie.Background, ie.BaseSrcRect, ie.BaseDstRect); err != nil {
		return err
	}

	if err := ie.Content.Render(view); err != nil {
		return err
	}

	if err := view.ScreenRenderer.Copy(ie.Border, ie.BaseSrcRect, ie.BaseDstRect); err != nil {
		return err
	}

	return nil
}

func (ie *InfoElement) GetBounds() *sdl.Rect {
	return ie.BaseDstRect
}

func (ie *InfoElement) ReplaceContent(newRender ViewElement) {
	if newRender == nil {
		logger.Debug("clearing info pane")
	} else {
		logger.Debug("setting info pane")
	}

	ie.Lock()
	defer ie.Unlock()
	ie.Content = newRender
}
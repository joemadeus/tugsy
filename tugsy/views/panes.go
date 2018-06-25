package views

import (
	"sync"

	"github.com/joemadeus/tugsy/tugsy/config"
	image "github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	infoBackgroundFile = "infoBackground.png"
	infoBorderFile     = "infoBorder.png"

	infoPaneDstX = 230
	infoPaneDstY = 670
	infoPaneSrcX = 0
	infoPaneSrcY = 0
	infoPaneH    = 256
	infoPaneW    = 192
)

// A PaneRenderStyle has a base graphic with static size and position, onto/into which
// other graphics are overlaid
type PaneRenderStyle struct {
	sync.Mutex

	BackgroundTexture *sdl.Texture
	CurrentRender     Render
	ForegroundBorder  *sdl.Texture

	BaseSrcRect *sdl.Rect
	BaseDstRect *sdl.Rect
}

func (style *PaneRenderStyle) Render(view *View) error {
	var err error
	if style.BackgroundTexture != nil {
		if err = view.ScreenRenderer.Copy(style.BackgroundTexture, style.BaseSrcRect, style.BaseDstRect); err != nil {
			return err
		}
	}

	style.Lock()
	if err = style.CurrentRender.Render(view); err != nil {
		return err
	}
	style.Unlock()

	if style.ForegroundBorder != nil {
		if err = view.ScreenRenderer.Copy(style.ForegroundBorder, style.BaseSrcRect, style.BaseDstRect); err != nil {
			return err
		}
	}

	return nil
}

func (style *PaneRenderStyle) GetBounds() *sdl.Rect {
	return style.BaseDstRect
}

func (style *PaneRenderStyle) ReplaceContent(newRender Render) {
	style.Lock()
	defer style.Unlock()
	style.CurrentRender = newRender
}

func NewInfoPaneRenderStyle(screenRenderer *sdl.Renderer, config *config.Config) (*PaneRenderStyle, error) {
	logger.Info("Loading info pane style")
	backgroundTex, err := image.LoadTexture(screenRenderer, config.GetSpritesheetPath(infoBackgroundFile))
	if err != nil {
		return nil, err
	}

	borderTex, err := image.LoadTexture(screenRenderer, config.GetSpritesheetPath(infoBorderFile))
	if err != nil {
		return nil, err
	}

	return &PaneRenderStyle{
		BackgroundTexture: backgroundTex,
		ForegroundBorder:  borderTex,

		BaseSrcRect:   &sdl.Rect{X: infoPaneSrcX, Y: infoPaneSrcY, H: infoPaneH, W: infoPaneW},
		BaseDstRect:   &sdl.Rect{X: infoPaneDstX, Y: infoPaneDstY, H: infoPaneH, W: infoPaneW},
		CurrentRender: &EmptyRenderStyle{},
	}, nil
}

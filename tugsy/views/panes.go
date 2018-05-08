package views

import (
	"errors"
	"sync"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	infoPaneX = 230
	infoPaneY = 670
	infoPaneH = 120
	infoPaneW = 240
)

type PaneRenderStyle struct {
	sync.Mutex
	*PaneSheet
	Sheet   *SpriteSheet
	SrcRect *sdl.Rect
	DstRect *sdl.Rect
}

// An InfoPaneStyle creates the pane on which port, weather, ship and other
// information is presented
type InfoPaneStyle struct {
	*PaneRenderStyle
	CurrentlyDisplayed Render
}

func NewInfoPaneRenderStyle(startingDisplay Render, spriteSet *SpriteSet) (*InfoPaneStyle, error) {
	logger.Info("Loading info pane style")
	srcRect, ok := spriteSet.PaneSheet.getSprite("info")
	if ok == false {
		return nil, errors.New("couldn't load the info pane -- no 'info' sprite")
	}

	return &InfoPaneStyle{
		PaneRenderStyle: &PaneRenderStyle{
			PaneSheet: spriteSet.PaneSheet,
			SrcRect:   srcRect,
			DstRect:   &sdl.Rect{X: infoPaneX, Y: infoPaneY, H: infoPaneH, W: infoPaneW},
		},
		CurrentlyDisplayed: startingDisplay,
	}, nil
}

func (style *InfoPaneStyle) Render(view *View) error {
	return view.ScreenRenderer.Copy(style.Sheet.Texture, style.SrcRect, style.DstRect)
}

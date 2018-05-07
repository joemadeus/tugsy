package views

import (
	"errors"
	"sync"

	"github.com/veandco/go-sdl2/sdl"
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
type InfoPaneStyle PaneRenderStyle

func NewInfoPaneRenderStyle(x, y, h, w int32, paneSheet *PaneSheet) (*InfoPaneStyle, error) {
	logger.Info("Loading info pane styles")
	srcRect, ok := paneSheet.getSprite("info")
	if ok == false {
		return nil, errors.New("couldn't load the info pane -- no 'info' sprite")
	}

	return &InfoPaneStyle{
		PaneSheet: paneSheet,
		SrcRect:   srcRect,
		DstRect:   &sdl.Rect{x, y, w, h},
	}, nil
}

func (style *InfoPaneStyle) Render(view *View) error {
	return view.ScreenRenderer.Copy(style.Sheet.Texture, style.SrcRect, style.DstRect)
}

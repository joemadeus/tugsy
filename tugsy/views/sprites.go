package views

import (
	"github.com/joemadeus/tugsy/tugsy/config"
	image "github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	defaultSpriteSizePixels = 40
	dotsSpritesFile         = "dots-normal.png"
	specialSpritesFile      = "special.png"
	flagsSpritesFile        = "flags.png"
	panesSpritesFile        = "panes.png"
)

type SpriteSet struct {
	DotSheet     *DotSheet
	SpecialSheet *SpecialSheet
	FlagSheet    *FlagSheet
	PaneSheet    *PaneSheet
}

func NewSpriteSet(screenRenderer *sdl.Renderer, config *config.Config) (*SpriteSet, error) {
	dots, err := NewDotSheet(screenRenderer, config)
	if err != nil {
		logger.Fatal("could not init the dots sprites", "error", err)
		return nil, err
	}

	special, err := NewSpecialSheet(screenRenderer, config)
	if err != nil {
		logger.Fatal("could not init the special sprites", "error", err)
		return nil, err
	}

	//flags, err := NewFlagSheet(screenRenderer, config)
	//if err != nil {
	//	logger.Fatal("could not init the flags sprites", "error", err)
	//	return nil, err
	//}

	panes, err := NewPaneSheet(screenRenderer, config)
	if err != nil {
		logger.Fatal("could not init the panes sprites", "error", err)
		return nil, err
	}

	return &SpriteSet{
		DotSheet:     dots,
		SpecialSheet: special,
		PaneSheet:    panes,
	}, nil
}

type SpriteSheet struct {
	*sdl.Texture
	SpriteSize int32
}

func (sheet *SpriteSheet) getSourceRect(row, column int32) *sdl.Rect {
	return &sdl.Rect{
		H: sheet.SpriteSize,
		W: sheet.SpriteSize,
		X: column * sheet.SpriteSize,
		Y: row * sheet.SpriteSize,
	}
}

type DotSheet struct {
	SpriteSheet
	DotMap      map[Hue]int32    // a dot hue to its row number, zero based
	ModifierMap map[string]int32 // a "modifier" string name to its column
}

func NewDotSheet(screenRenderer *sdl.Renderer, config *config.Config) (*DotSheet, error) {
	logger.Info("Loading sprites \"Dots\"")
	tex, err := image.LoadTexture(screenRenderer, config.GetSpritesheetPath(dotsSpritesFile))
	if err != nil {
		return nil, err
	}

	dots := &DotSheet{}
	dots.Texture = tex
	dots.SpriteSize = defaultSpriteSizePixels

	dots.ModifierMap = make(map[string]int32)
	dots.ModifierMap["normal"] = 0
	dots.ModifierMap["lighter"] = 1

	dots.DotMap = make(map[Hue]int32)
	i := 0
	for i <= 18 {
		dots.DotMap[Hue(i*20+10)] = int32(i)
		i += 1
	}

	return dots, nil
}

func (dots *DotSheet) Teardown() {
	dots.Texture.Destroy()
}

func (dots *DotSheet) GetSprite(hue Hue, modifier string) (*sdl.Rect, bool) {
	row, ok := dots.DotMap[hue]
	if ok == false {
		logger.Warn("hue is unknown", "hue", hue)
		return nil, false
	}

	column, ok := dots.ModifierMap[modifier]
	if ok == false {
		logger.Warn("modifier is unknown", "modifier", modifier)
		return nil, false
	}

	return dots.getSourceRect(row, column), true
}

type SpecialSheet struct {
	SpriteSheet
	MarkerMap map[string]int32 // the name of the sprite to its row number, zero based
}

func NewSpecialSheet(screenRenderer *sdl.Renderer, config *config.Config) (*SpecialSheet, error) {
	logger.Info("Loading sprites \"Special\"")
	tex, err := image.LoadTexture(screenRenderer, config.GetSpritesheetPath(specialSpritesFile))
	if err != nil {
		return nil, err
	}

	special := &SpecialSheet{}
	special.Texture = tex
	special.SpriteSize = defaultSpriteSizePixels

	special.MarkerMap = make(map[string]int32)
	special.MarkerMap["unknown"] = int32(0)
	special.MarkerMap["hazard_a"] = int32(1)
	special.MarkerMap["hazard_b"] = int32(2)
	special.MarkerMap["hazard_c"] = int32(3)
	special.MarkerMap["hazard_d"] = int32(4)

	return special, nil
}

func (special *SpecialSheet) Teardown() {
	special.Texture.Destroy()
}

func (special *SpecialSheet) GetSprite(spriteName string) (*sdl.Rect, bool) {
	row, ok := special.MarkerMap[spriteName]
	if ok == false {
		logger.Warn("special: sprite is unknown", "spriteName", spriteName)
		return nil, false
	}

	return special.getSourceRect(row, 0), true
}

type FlagSheet struct {
	*SpriteSheet
	FlagMap map[string][2]uint8 // country code (ISO 3166-1 alpha-2) to flag X:Y coordinate
}

func NewFlagSheet(screenRenderer *sdl.Renderer, config *config.Config) (*FlagSheet, error) {
	logger.Info("Loading sprites \"Flags\"")

	tex, err := image.LoadTexture(screenRenderer, config.GetSpritesheetPath(flagsSpritesFile))
	if err != nil {
		return nil, err
	}

	flags := &FlagSheet{}
	flags.Texture = tex

	return flags, nil
}

type PaneSheet struct {
	SpriteSheet
	PaneMap map[string]int32
}

func NewPaneSheet(screenRenderer *sdl.Renderer, config *config.Config) (*PaneSheet, error) {
	logger.Info("Loading sprites \"Panes\"")

	tex, err := image.LoadTexture(screenRenderer, config.GetSpritesheetPath(panesSpritesFile))
	if err != nil {
		return nil, err
	}

	panes := &PaneSheet{}
	panes.Texture = tex
	panes.SpriteSize = 128

	panes.PaneMap = make(map[string]int32)
	panes.PaneMap["info"] = int32(0)
	panes.PaneMap["tidebar"] = int32(1)
	panes.PaneMap["skycolor"] = int32(2)
	panes.PaneMap["wx_hazard"] = int32(3)

	return panes, nil
}

func (panes *PaneSheet) getSprite(spriteName string) (*sdl.Rect, bool) {
	row, ok := panes.PaneMap[spriteName]
	if ok == false {
		return nil, false
	}
	return panes.getSourceRect(row, 0), true
}

package views

import (
	"errors"

	"github.com/joemadeus/tugsy/tugsy/config"
	logger "github.com/sirupsen/logrus"
	image "github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	defaultSpriteSizePixels = 40

	dotsSpritesFile    = "dots-normal.png"
	specialSpritesFile = "special.png"
	flagsSpritesFile   = "flags.png"
)

var (
	UnknownSpriteErr   = errors.New("unknown sprite")
	UnknownModifierErr = errors.New("unknown sprite modifier")
)

type SpriteSet struct {
	DotSheet     *DotSheet
	SpecialSheet *SpecialSheet
	FlagSheet    *FlagSheet
}

func NewSpriteSet(screenRenderer *sdl.Renderer, config *config.Config) (*SpriteSet, error) {
	dots, err := NewDotSheet(screenRenderer, config)
	if err != nil {
		logger.WithError(err).Fatal("could not init the dots sprites")
		return nil, err
	}

	special, err := NewSpecialSheet(screenRenderer, config)
	if err != nil {
		logger.WithError(err).Fatal("could not init the special sprites")
		return nil, err
	}

	// flags, err := NewFlagSheet(screenRenderer, config)
	// if err != nil {
	// 	logger.WithError(err).Fatal("could not init the flags sprites")
	// 	return nil, err
	// }

	return &SpriteSet{
		DotSheet:     dots,
		SpecialSheet: special,
	}, nil
}

type Teardownable interface {
	Teardown() error
}

func sourceRect(row, column, size int32) *sdl.Rect {
	return &sdl.Rect{
		H: size,
		W: size,
		X: column * size,
		Y: row * size,
	}
}

type Sprite struct {
	*sdl.Rect
	*sdl.Texture
}

type DotSheet struct {
	*sdl.Texture
	SpriteSize  int32
	DotMap      map[Hue]int32    // a dot hue to its row number, zero based
	ModifierMap map[string]int32 // a "modifier" string name to its column
}

func NewDotSheet(screenRenderer *sdl.Renderer, config *config.Config) (*DotSheet, error) {
	logger.Info("Loading sprites 'Dots'")
	tex, err := image.LoadTexture(screenRenderer, config.SpriteSheetPath(dotsSpritesFile))
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

func (dots *DotSheet) Teardown() error {
	if err := dots.Texture.Destroy(); err != nil {
		logger.WithError(err).Error("while tearing down 'dots' sprite sheet")
		return err
	}

	return nil
}

func (dots *DotSheet) GetSprite(hue Hue, modifier string) (*Sprite, error) {
	row, ok := dots.DotMap[hue]
	if ok == false {
		return nil, UnknownSpriteErr
	}

	column, ok := dots.ModifierMap[modifier]
	if ok == false {
		return nil, UnknownModifierErr
	}

	return &Sprite{
		Texture: dots.Texture,
		Rect:    sourceRect(row, column, dots.SpriteSize),
	}, nil
}

type SpecialSheet struct {
	*sdl.Texture
	SpriteSize int32
	MarkerMap  map[string]int32 // the name of the sprite to its row number, zero based
}

func NewSpecialSheet(screenRenderer *sdl.Renderer, config *config.Config) (*SpecialSheet, error) {
	logger.Info("Loading sprites 'Special'")
	tex, err := image.LoadTexture(screenRenderer, config.SpriteSheetPath(specialSpritesFile))
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

func (special *SpecialSheet) Teardown() error {
	if err := special.Texture.Destroy(); err != nil {
		logger.WithError(err).Error("while tearing down 'special' sprite sheet")
		return err
	}

	return nil
}

func (special *SpecialSheet) GetSprite(spriteName string) (*Sprite, error) {
	row, ok := special.MarkerMap[spriteName]
	if ok == false {
		return nil, UnknownSpriteErr
	}

	return &Sprite{
		Texture: special.Texture,
		Rect:    sourceRect(row, 0, special.SpriteSize),
	}, nil
}

type FlagSheet struct {
	*sdl.Texture
	SpriteSize int32
	FlagMap    map[string][2]uint8 // country code (ISO 3166-1 alpha-2) to flag X:Y coordinate
}

func NewFlagSheet(screenRenderer *sdl.Renderer, config *config.Config) (*FlagSheet, error) {
	logger.Info("Loading sprites 'Flags'")

	tex, err := image.LoadTexture(screenRenderer, config.SpriteSheetPath(flagsSpritesFile))
	if err != nil {
		return nil, err
	}

	flags := &FlagSheet{}
	flags.Texture = tex

	return flags, nil
}

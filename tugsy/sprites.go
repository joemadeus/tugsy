package main

import (
	image "github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

// Returns a path to a resource in the given view
func getSpritePath(pngResource string) string {
	return spritesDir + "/" + pngResource
}

type SpriteSheet struct {
	*sdl.Texture
	TextureSource string
	SpriteSize    int32
}

func (sheet *SpriteSheet) getSprite() {}

type Dots struct {
	SpriteSheet
	DotMap     map[uint8]int32      // a dot hue to its row number, zero based
	SpecialMap map[string]int32     // a "special" string to its column
}

func NewDots(screenRenderer *sdl.Renderer, config *Config, spritesDir string) (*Dots, error) {
	dots := &Dots{}
	err := config.UnmarshalKey("sprites.dots", dots)
	if err != nil {
		return nil, err
	}

	dots.Texture, err = image.LoadTexture(screenRenderer, spritesDir+"/"+dots.TextureSource)
	if err != nil {
		return nil, err
	}

	// TODO: Unfortunate mix of configuration and programmatic evaluation, here

	dots.SpecialMap = make(map[string]uint8)
	dots.SpecialMap["normal"] = 0
	dots.SpecialMap["lighter"] = 1

	dots.DotMap = make(map[uint8]uint8)
	i := 0
	for i <= 18 {
		dots.DotMap[uint8(i*20+10)] = uint8(i)
		i += 1
	}

	return dots, nil
}

func (dots *Dots) getSourceRect(hue uint8, special string) *sdl.Rect {
	x, ok := dots.SpecialMap[special]
	if ok == false {
		x = 0 // default to a "normal" dot
	}
	x *= dots.SpriteSize


	return &sdl.Rect{
		H: dots.SpriteSize,
		W: dots.SpriteSize,
		X: x,
	}
}

type Special struct {
	SpriteSheet
	MarkerMap map[string]uint8 // the name of the sprite to its row number, zero based
}

func NewSpecial(screenRenderer *sdl.Renderer, config *Config, spritesDir string) (*Special, error) {
	special := &Special{}
	err := config.UnmarshalKey("sprites.dots", special)
	if err != nil {
		return nil, err
	}

	special.Texture, err = image.LoadTexture(screenRenderer, spritesDir+"/"+special.TextureSource)
	if err != nil {
		return nil, err
	}

	// TODO: could be set via config, instead

	special.MarkerMap = make(map[string]uint8)
	special.MarkerMap["hazard_a"] = uint8(0)
	special.MarkerMap["hazard_b"] = uint8(1)
	special.MarkerMap["hazard_c"] = uint8(2)
	special.MarkerMap["hazard_d"] = uint8(3)
	special.MarkerMap["unknown"] = uint8(4)

	return special, nil
}

type Flags struct {
	*SpriteSheet
	FlagMap map[uint8][2]uint8 // a flag ID to its X:Y coordinate
}

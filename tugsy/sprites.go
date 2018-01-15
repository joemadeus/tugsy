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

func (sheet *SpriteSheet) getSourceRect(row, column int32) *sdl.Rect {
	return &sdl.Rect{
		H: sheet.SpriteSize,
		W: sheet.SpriteSize,
		X: column * sheet.SpriteSize,
		Y: row * sheet.SpriteSize,
	}
}

type Dots struct {
	SpriteSheet
	DotMap      map[Hue]int32  // a dot hue to its row number, zero based
	ModifierMap map[string]int32 // a "modifier" string name to its column
}

func NewDots(screenRenderer *sdl.Renderer) (*Dots, error) {
	dots := &Dots{}
	var err error
	dots.Texture, err = image.LoadTexture(screenRenderer, getSpritePath(dots.TextureSource))
	if err != nil {
		return nil, err
	}

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

func (dots *Dots) GetSprite(hue Hue, modifier string) (*sdl.Rect, *SpriteSheet, bool) {
	row, ok := dots.DotMap[hue]
	if ok == false {
		logger.Warn("Hue is unknown", "Hue", hue)
		return nil, nil, false
	}

	column, ok := dots.ModifierMap[modifier]
	if ok == false {
		logger.Warn("Modifier is unknown", "modifier name", modifier)
		return nil, nil, false
	}

	return dots.getSourceRect(row, column), &dots.SpriteSheet, true
}

type Special struct {
	SpriteSheet
	MarkerMap map[string]int32 // the name of the sprite to its row number, zero based
}

func NewSpecial(screenRenderer *sdl.Renderer) (*Special, error) {
	special := &Special{}
	var err error
	special.Texture, err = image.LoadTexture(screenRenderer, getSpritePath(special.TextureSource))
	if err != nil {
		return nil, err
	}

	// TODO: could be set via config, instead

	special.MarkerMap = make(map[string]int32)
	special.MarkerMap["unknown"] = int32(0)
	special.MarkerMap["hazard_a"] = int32(1)
	special.MarkerMap["hazard_b"] = int32(2)
	special.MarkerMap["hazard_c"] = int32(3)
	special.MarkerMap["hazard_d"] = int32(4)

	return special, nil
}

func (special *Special) GetSprite(spriteName string) (*sdl.Rect, *SpriteSheet, bool) {
	row, ok := special.MarkerMap[spriteName]
	if ok == false {
		return nil, nil, false
	}

	return special.getSourceRect(row, 0), &special.SpriteSheet, true
}

type Flags struct {
	*SpriteSheet
	FlagMap map[uint8][2]uint8 // a flag ID to its X:Y coordinate
}

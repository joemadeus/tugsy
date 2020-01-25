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

	flags, err := NewFlagSheet(screenRenderer, config)
	if err != nil {
		logger.WithError(err).Fatal("could not init the flags sprites")
		return nil, err
	}

	return &SpriteSet{
		DotSheet:     dots,
		SpecialSheet: special,
		FlagSheet:    flags,
	}, nil
}

func sourceRect(row, column int, size int32) *sdl.Rect {
	return &sdl.Rect{
		H: size,
		W: size,
		X: int32(column) * size,
		Y: int32(row) * size,
	}
}

type Sprite struct {
	*sdl.Rect
	*sdl.Texture
}

type DotSheet struct {
	*sdl.Texture
	SpriteSize  int32
	DotMap      map[Hue]int    // a dot hue to its row number, zero based
	ModifierMap map[string]int // a "modifier" string name to its column
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

	dots.ModifierMap = make(map[string]int)
	dots.ModifierMap["normal"] = 0
	dots.ModifierMap["lighter"] = 1

	dots.DotMap = make(map[Hue]int)
	i := 0
	for i <= 18 {
		dots.DotMap[Hue(i*20+10)] = int(i)
		i += 1
	}

	return dots, nil
}

func (d *DotSheet) Teardown() error {
	if err := d.Texture.Destroy(); err != nil {
		logger.WithError(err).Error("while tearing down 'dots' sprite sheet")
		return err
	}

	return nil
}

func (d *DotSheet) GetSprite(hue Hue, modifier string) (*Sprite, error) {
	row, ok := d.DotMap[hue]
	if ok == false {
		return nil, UnknownSpriteErr
	}

	column, ok := d.ModifierMap[modifier]
	if ok == false {
		return nil, UnknownModifierErr
	}

	return &Sprite{
		Texture: d.Texture,
		Rect:    sourceRect(row, column, d.SpriteSize),
	}, nil
}

type SpecialSheet struct {
	*sdl.Texture
	SpriteSize int32
	MarkerMap  map[string]int // the name of the sprite to its row number, zero based
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

	special.MarkerMap = make(map[string]int)
	special.MarkerMap["unknown"] = 0
	special.MarkerMap["hazard_a"] = 1
	special.MarkerMap["hazard_b"] = 2
	special.MarkerMap["hazard_c"] = 3
	special.MarkerMap["hazard_d"] = 4
	special.MarkerMap["red_ring"] = 5

	return special, nil
}

func (s *SpecialSheet) Teardown() error {
	if err := s.Texture.Destroy(); err != nil {
		logger.WithError(err).Error("while tearing down 'special' sprite sheet")
		return err
	}

	return nil
}

func (s *SpecialSheet) GetSprite(name string) (*Sprite, error) {
	row, ok := s.MarkerMap[name]
	if ok == false {
		return nil, UnknownSpriteErr
	}

	return &Sprite{
		Texture: s.Texture,
		Rect:    sourceRect(row, 0, s.SpriteSize),
	}, nil
}

type FlagSheet struct {
	*sdl.Texture
	SpriteSize int32
	FlagMap    map[string]int // country code (ISO 3166-1 alpha-2) to flag row
}

func NewFlagSheet(screenRenderer *sdl.Renderer, config *config.Config) (*FlagSheet, error) {
	logger.Info("Loading sprites 'Flags'")

	tex, err := image.LoadTexture(screenRenderer, config.SpriteSheetPath(flagsSpritesFile))
	if err != nil {
		return nil, err
	}

	flags := &FlagSheet{}
	flags.Texture = tex
	flags.SpriteSize = 64
	flags.FlagMap = flagMap()

	return flags, nil
}

func (f *FlagSheet) GetSprite(iso string) (*Sprite, error) {
	row, ok := f.FlagMap[iso]
	if ok == false {
		return nil, UnknownSpriteErr
	}

	return &Sprite{
		Texture: f.Texture,
		Rect:    sourceRect(row, 0, f.SpriteSize),
	}, nil
}

func (f *FlagSheet) Teardown() error {
	if err := f.Texture.Destroy(); err != nil {
		logger.WithError(err).Error("while tearing down 'flags' sprite sheet")
		return err
	}

	return nil
}

func flagMap() map[string]int {
	ret := make(map[string]int)
	for i, iso := range []string{
		"AD", "AE", "AF", "AG", "AI", "AL", "AM", "AN", "AO", "AQ", "AR", "AS",
		"AT", "AU", "AW", "AX", "AZ", "BA", "BB", "BD", "BE", "BF", "BG", "BH",
		"BI", "BJ", "BL", "BM", "BN", "BO", "BR", "BS", "BT", "BW", "BY", "BZ",
		"CA", "CC", "CD", "CF", "CG", "CH", "CI", "CK", "CL", "CM", "CN", "CO",
		"CR", "CU", "CV", "CW", "CX", "CY", "CZ", "DE", "DJ", "DK", "DM", "DO",
		"DZ", "EC", "EE", "EG", "EH", "ER", "ES", "ET", "EU", "FI", "FJ", "FK",
		"FM", "FO", "FR", "GA", "GB", "GD", "GE", "GG", "GH", "GI", "GL", "GM",
		"GN", "GQ", "GR", "GS", "GT", "GU", "GW", "GY", "HK", "HN", "HR", "HT",
		"HU", "IC", "ID", "IE", "IL", "IM", "IN", "IQ", "IR", "IS", "IT", "JE",
		"JM", "JO", "JP", "KE", "KG", "KH", "KI", "KM", "KN", "KP", "KR", "KW",
		"KY", "KZ", "LA", "LB", "LC", "LI", "LK", "LR", "LS", "LT", "LU", "LV",
		"LY", "MA", "MC", "MD", "ME", "MF", "MG", "MH", "MK", "ML", "MM", "MN",
		"MO", "MP", "MQ", "MR", "MS", "MT", "MU", "MV", "MW", "MX", "MY", "MZ",
		"NA", "NC", "NE", "NF", "NG", "NI", "NL", "NO", "NP", "NR", "NU", "NZ",
		"OM", "PA", "PE", "PF", "PG", "PH", "PK", "PL", "PN", "PR", "PS", "PT",
		"PW", "PY", "QA", "RO", "RS", "RU", "RW", "SA", "SB", "SC", "SD", "SE",
		"SG", "SH", "SI", "SK", "SL", "SM", "SN", "SO", "SR", "SS", "ST", "SV",
		"SY", "SZ", "TC", "TD", "TF", "TG", "TH", "TJ", "TK", "TL", "TM", "TN",
		"TO", "TR", "TT", "TV", "TW", "TZ", "UA", "UG", "US", "UY", "UZ", "VA",
		"VC", "VE", "VG", "VI", "VN", "VU", "WF", "WS", "YE", "YT", "ZA", "ZM",
		"ZW", "_abkhazia", "_adelie", "_azores", "_crozet", "_england",
		"_kerguelen", "_kosovo", "_madeira", "_palestine",
		"_reunion", "_scotland", "_south-ossetia", "_stpaul", "_wales",
	} {
		ret[iso] = i
	}

	return ret
}

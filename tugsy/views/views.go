package views

import (
	"errors"

	"github.com/andmarios/aislib"
	"github.com/joemadeus/tugsy/tugsy/config"
	logger "github.com/sirupsen/logrus"
	image "github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	baseMapFile  = "/base.png"
	ScreenWidth  = 480
	ScreenHeight = 760
	ScreenTitle  = "Tugsy"
)

var (
	NoViewConfigFound = errors.New("could not find view configs")
	PixelFormat       = uint32(sdl.PIXELFORMAT_RGBA32)
)

type Teardownable interface {
	Teardown() error
}

type ViewConfig struct {
	MapName string
	North   float64
	South   float64
	East    float64
	West    float64
}

type ViewSet struct {
	resourceDir string

	index int
	Views []*View
}

func ViewSetFromConfig(config *config.Config, screenRenderer *sdl.Renderer, re *RootElement) (*ViewSet, error) {
	if config.IsSet("views") == false {
		return nil, NoViewConfigFound
	}

	var viewConfigs []*ViewConfig
	if err := config.UnmarshalKey("views", &viewConfigs); err != nil {
		return nil, err
	}

	viewSet := &ViewSet{
		Views: make([]*View, 0),
	}

	for _, viewConfig := range viewConfigs {
		logger.Infof("Loading view %s", viewConfig.MapName)
		baseTexture, err := image.LoadTexture(screenRenderer, config.ViewPath(viewConfig.MapName)+baseMapFile)
		if err != nil {
			return nil, err
		}

		baseMap := &BaseMap{
			Tex:    baseTexture,
			SWGeo:  RealWorldPosition{viewConfig.West, viewConfig.South},
			NEGeo:  RealWorldPosition{viewConfig.East, viewConfig.North},
			Name:   viewConfig.MapName,
			width:  float64(ScreenWidth),
			height: float64(ScreenHeight),
		}

		view := &View{
			BaseMap:        baseMap,
			ScreenRenderer: screenRenderer,
			RootElement:    re,
		}

		viewSet.Views = append(viewSet.Views, view)
	}

	return viewSet, nil
}

func (vs *ViewSet) CurrentView() *View {
	return vs.Views[vs.index]
}

func (vs *ViewSet) NextView() *View {
	if vs.index == len(vs.Views)-1 {
		vs.index = 0
	} else {
		vs.index++
	}
	return vs.Views[vs.index]
}

func (vs *ViewSet) Teardown() error {
	logger.Info("Tearing down views")
	for _, view := range vs.Views {
		logger.Infof("Unloading view %s", view.Name)
		if err := view.BaseMap.Tex.Destroy(); err != nil {
			logger.WithError(err).Errorf("while tearing down view %s", view.Name)
			// TODO
		}
	}

	return nil
}

type View struct {
	*BaseMap
	ScreenRenderer *sdl.Renderer
	RootElement    *RootElement
}

// Display clears the renderer and redisplays all the ViewElements in this View
func (v *View) Display() error {
	err := v.ScreenRenderer.Clear()
	if err != nil {
		logger.WithError(err).Error("Could not clear the screen renderer")
		return err
	}

	err = v.ScreenRenderer.Copy(v.BaseMap.Tex, nil, nil)
	if err != nil {
		logger.WithError(err).Error("could not copy the base map to the screen renderer")
		return err
	}

	if err := v.RootElement.Render(v); err != nil {
		logger.WithError(err).Errorf("could not render")
	}

	v.ScreenRenderer.Present()

	return nil
}

// GetBaseMapPosition estimates the given position report on the view's base map
// using a simple linear approximation
func (v *View) BaseMapPosition(position *aislib.PositionReport) BaseMapPosition {
	return BaseMapPosition{
		(position.Lon - v.SWGeo.X) / (v.NEGeo.X - v.SWGeo.X) * v.width,
		v.height - (position.Lat-v.SWGeo.Y)/(v.NEGeo.Y-v.SWGeo.Y)*v.height,
	}
}

type position struct {
	X float64
	Y float64
}

// RealWorldPosition is a degree decimal formatted latitude/longitude
type RealWorldPosition position

// BaseMapPosition is a pixel coordinate on a base map
type BaseMapPosition position

type BaseMap struct {
	Tex           *sdl.Texture
	NEGeo         RealWorldPosition
	SWGeo         RealWorldPosition
	Name          string
	width, height float64
}

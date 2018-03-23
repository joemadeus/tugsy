package views

import (
	"errors"

	"github.com/andmarios/aislib"
	"github.com/joemadeus/tugsy/tugsy/config"
	"github.com/joemadeus/tugsy/tugsy/shipdata"
	image "github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	baseMapFile  = "/base.png"
	ScreenWidth  = 480
	ScreenHeight = 760
	ScreenTitle  = "Tugsy"
)

var NoViewConfigFound = errors.New("could not find view configs")

type Hue uint16

type ViewSet struct {
	resourceDir string

	index int
	Views []*View
}

type ViewConfig struct {
	MapName string
	North   float64
	South   float64
	East    float64
	West    float64
}

func ViewSetFromConfig(screenRenderer *sdl.Renderer, config *config.Config) (*ViewSet, error) {
	if config.IsSet("views") == false {
		return nil, NoViewConfigFound
	}

	var viewConfigs []*ViewConfig
	err := config.UnmarshalKey("views", &viewConfigs)
	if err != nil {
		return nil, err
	}

	viewSet := &ViewSet{
		index: 0,
		Views: make([]*View, 0),
	}

	currentPositionRenderer, err := shipdata.NewCurrentPositionByType(screenRenderer, config)
	if err != nil {
		return nil, err
	}

	markPathRenderer, err := shipdata.NewMarkPathByType()
	if err != nil {
		return nil, err
	}

	for _, viewConfig := range viewConfigs {
		logger.Info("Loading", "viewName", viewConfig.MapName)
		baseTexture, err := image.LoadTexture(screenRenderer, config.GetViewPath(viewConfig.MapName)+baseMapFile)
		if err != nil {
			return nil, err
		}

		baseMap := &BaseMap{
			Tex:    baseTexture,
			SWGeo:  RealWorldPosition{viewConfig.West, viewConfig.South},
			NEGeo:  RealWorldPosition{viewConfig.East, viewConfig.North},
			width:  float64(ScreenWidth),
			height: float64(ScreenHeight),
		}

		view := &View{
			BaseMap:        baseMap,
			ViewName:       viewConfig.MapName,
			ScreenRenderer: screenRenderer,
		}

		view.renderPathFunc = markPathRenderer.Render(view)
		view.renderCurrentPositionFunc = currentPositionRenderer.Render(view)

		viewSet.Views = append(viewSet.Views, view)
	}

	return viewSet, nil
}

func (viewSet *ViewSet) CurrentView() *View {
	return viewSet.Views[viewSet.index]
}

func (viewSet *ViewSet) NextView() *View {
	if viewSet.index == len(viewSet.Views)-1 {
		viewSet.index = 0
	} else {
		viewSet.index += 1
	}
	return viewSet.Views[viewSet.index]
}

func (viewSet *ViewSet) TeardownResources() {
	logger.Info("Tearing down views")
	for _, view := range viewSet.Views {
		logger.Info("Unloading view", view.ViewName)
		view.BaseMap.Tex.Destroy()
	}
}

type View struct {
	*BaseMap
	ViewName       string
	ScreenRenderer *sdl.Renderer

	renderPathFunc, renderCurrentPositionFunc shipdata.ShipDataRenderFunction
}

// Clears the renderer and redisplays the base map and all tracks
func (view *View) Display() error {
	err := view.ScreenRenderer.Clear()
	if err != nil {
		logger.Warn("Could not clear the screen renderer", "err", err)
		return err
	}

	err = view.ScreenRenderer.Copy(view.BaseMap.Tex, nil, nil)
	if err != nil {
		logger.Warn("Could not copy the base map to the screen renderer", "err", err)
		return err
	}

	shipdata.PositionData.RenderPositionReports(view.renderPathFunc, view.renderCurrentPositionFunc)

	view.ScreenRenderer.Present()

	return nil
}

// getBaseMapPosition returns the given real world position on the view's
// base map using a simple linear approximation
func (view *View) GetBaseMapPosition(position *aislib.PositionReport) BaseMapPosition {
	return BaseMapPosition{
		(position.Lon - view.SWGeo.X) / (view.NEGeo.X - view.SWGeo.X) * view.width,
		view.height - (position.Lat-view.SWGeo.Y)/(view.NEGeo.Y-view.SWGeo.Y)*view.height,
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
	width, height float64
}

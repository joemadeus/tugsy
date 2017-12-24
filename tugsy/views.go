package main

import (
	"errors"

	"fmt"

	"github.com/andmarios/aislib"
	image "github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	baseMapFile = "/base.png"

	screenWidth  = 480
	screenHeight = 760
	screenTitle  = "Tugsy"
)

var NoViewConfigFound = errors.New("could not find view configs")

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

func ViewSetFromConfig(screenRenderer *sdl.Renderer, config *Config) (*ViewSet, error) {
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

	for _, viewConfig := range viewConfigs {
		logger.Info("Loading", "viewName", viewConfig.MapName)
		baseTexture, err := image.LoadTexture(screenRenderer, viewSet.getResourcePath(viewConfig.MapName, baseMapFile))
		if err != nil {
			return nil, err
		}

		baseMap := &BaseMap{
			Tex:    baseTexture,
			SWGeo:  RealWorldPosition{viewConfig.West, viewConfig.South},
			NEGeo:  RealWorldPosition{viewConfig.East, viewConfig.North},
			width:  float64(screenWidth),
			height: float64(screenHeight),
		}

		viewSet.Views = append(viewSet.Views, &View{
			BaseMap:                   baseMap,
			ViewName:                  viewConfig.MapName,
			screenRenderer:            screenRenderer,
			renderCurrentPositionFunc: &CurrentPositionSimple{},
			renderPathFunc:            &MarkPathSimple{},
		})
	}

	return viewSet, nil
}

// Returns a path to a resource in the given view
func (viewSet *ViewSet) getResourcePath(viewName string, pngResource string) string {
	return viewSet.resourceDir + "/" + viewName + pngResource
}

func (viewSet *ViewSet) currentView() *View {
	return viewSet.Views[viewSet.index]
}

func (viewSet *ViewSet) nextView() *View {
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
	screenRenderer *sdl.Renderer

	renderPathFunc, renderCurrentPositionFunc RenderStyle
}

// Clears the renderer and redisplays the base map and all tracks
func (view *View) Display() error {
	if MachineAndProcessState.TheData.dirty == false {
		// if there are no updates since the last refresh, don't
		// bother redisplaying
		return nil
	}

	err := view.screenRenderer.Clear()
	if err != nil {
		logger.Warn("Could not clear the screen renderer", "err", err)
		return err
	}

	err = view.screenRenderer.Copy(view.BaseMap.Tex, nil, nil)
	if err != nil {
		logger.Warn("Could not copy the base map to the screen renderer", "err", err)
		return err
	}

	for _, mmsi := range MachineAndProcessState.TheData.GetHistoryMMSIs() {
		ok := MachineAndProcessState.TheData.RenderPositionReports(
			mmsi,
			view.renderPathFunc.Render(view),
			view.renderCurrentPositionFunc.Render(view))

		if ok == false {
			logger.Info("A ShipData has no points or was removed", "mmsi", fmt.Sprintf("%d", mmsi))
			continue
		}
	}

	view.screenRenderer.Present()

	return nil
}

// getBaseMapPosition returns the given real world position on the view's
// base map using a simple linear approximation
func (view *View) getBaseMapPosition(position *aislib.PositionReport) BaseMapPosition {
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

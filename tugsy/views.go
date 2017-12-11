package main

import (
	"errors"

	image "github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	resDir      = "./res"
	baseMapFile = "/base.png"

	screenWidth  = 480
	screenHeight = 600
	screenTitle  = "Tugsy"

	trackLinesR = 192
	trackLinesG = 192
	trackLinesB = 0

	trackPointsR = 128
	trackPointsG = 128
	trackPointsB = 0
)

var NoViewConfigFound = errors.New("could not find view configs")

// Returns a path to a resource in the given view
func getResourcePath(viewName string, pngResource string) string {
	return resDir + "/" + viewName + pngResource
}

type ViewSet struct {
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
		baseTexture, err := image.LoadTexture(screenRenderer, getResourcePath(viewConfig.MapName, baseMapFile))
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
			BaseMap:        baseMap,
			ViewName:       viewConfig.MapName,
			screenRenderer: screenRenderer,
		})
	}

	return viewSet, nil
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
		var currentPosition sdl.Point
		var sdlPoints []sdl.Point
		translatePositionFunc := func(positionReports []Positionable) {
			sdlPoints := make([]sdl.Point, 0, len(positionReports))
			for i, positionReport := range positionReports {
				realWorldPosition := RealWorldPosition{
					X: positionReport.GetPositionReport().Lat,
					Y: positionReport.GetPositionReport().Lon,
				}
				baseMapPosition := view.getBaseMapPosition(realWorldPosition)
				sdlPoints[i] = sdl.Point{
					X: int32(baseMapPosition.X + 0.5),
					Y: int32(baseMapPosition.Y + 0.5),
				}
				currentPosition = sdlPoints[i]
			}
		}

		ok := MachineAndProcessState.TheData.TranslatePositionReports(mmsi, translatePositionFunc)
		if ok == false {
			logger.Info("A ShipData was removed before we could render it", "mmsi", mmsi)
			continue
		}

		view.screenRenderer.SetDrawColor(trackLinesR, trackLinesG, trackLinesB, sdl.ALPHA_OPAQUE)
		view.screenRenderer.DrawLines(sdlPoints)

		view.screenRenderer.SetDrawColor(trackPointsR, trackPointsG, trackPointsB, sdl.ALPHA_OPAQUE)
		view.screenRenderer.DrawPoints(sdlPoints)
		view.screenRenderer.DrawPoint(int(currentPosition.X), int(currentPosition.Y))
	}

	view.screenRenderer.Present()

	return nil
}

// getBaseMapPosition returns the given real world position on the view's
// base map using a simple linear approximation
func (view *View) getBaseMapPosition(position RealWorldPosition) BaseMapPosition {
	return BaseMapPosition{
		(position.X - view.SWGeo.X) / (view.NEGeo.X - view.SWGeo.X) * view.width,
		(position.Y - view.SWGeo.Y) / (view.NEGeo.Y - view.SWGeo.Y) * view.height,
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

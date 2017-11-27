package main

import (
	image "github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	resDir      = "./res"
	baseMapFile = "/base.png"
	spriteFile  = "/sprites.png"

	spriteSize = 8 // pixels square

	screenWidth  = 480
	screenHeight = 600
	screenTitle  = "Tugsy"

	targetFPS uint32 = 60

	trackLinesR = 192
	trackLinesG = 192
	trackLinesB = 0

	trackPointsR = 128
	trackPointsG = 128
	trackPointsB = 0
)

var ViewList = [3]string{
	"pvd_harbor",
	"pvd_to_bristol",
	"pvd_to_gansett",
}

var index = 0
var Views [len(ViewList)]*View

func currentView() *View {
	return Views[index]
}

func nextView() *View {
	if index == len(Views)-1 {
		index = 0
	} else {
		index += 1
	}
	return Views[index]
}

// Returns a path to a resource in the given view
func getResourcePath(viewName string, pngResource string) string {
	return resDir + "/" + viewName + pngResource
}

func InitResources(screenRenderer *sdl.Renderer) error {
	for index := range ViewList {
		viewName := ViewList[index]
		logger.Info("Loading view", "viewName", viewName)

		baseTexture, err := image.LoadTexture(screenRenderer, getResourcePath(ViewList[index], baseMapFile))
		if err != nil {
			return err
		}

		spriteTexture, err := image.LoadTexture(screenRenderer, getResourcePath(ViewList[index], spriteFile))
		if err != nil {
			return err
		}

		baseMap := &BaseMap{
			Tex:    baseTexture,
			LRGeo:  RealWorldPosition{0.0, 0.0},
			ULGeo:  RealWorldPosition{0.0, 0.0},
			width:  float64(screenWidth),
			height: float64(screenHeight),
		}
		sprites := &SpriteSheet{spriteTexture, spriteSize}
		Views[index] = &View{
			BaseMap:        baseMap,
			ViewName:       ViewList[index],
			Sprites:        sprites,
			screenRenderer: screenRenderer,
		}
	}

	return nil
}

func TeardownResources() {
	logger.Info("Tearing down views")
	for index := range ViewList {
		view := Views[index]
		logger.Info("Unloading view", view.ViewName)
		view.BaseMap.Tex.Destroy()
		view.Sprites.Tex.Destroy()
	}

}

type View struct {
	*BaseMap
	ViewName       string
	Sprites        *SpriteSheet
	screenRenderer *sdl.Renderer
}

// Clears the renderer and redisplays the base map and tracks
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
		positionReports := MachineAndProcessState.TheData.GetPositionReports(mmsi, false)
		if positionReports == nil {
			logger.Debug("An MMSI was removed before we could display it", "MMSI", mmsi)
			continue
		}

		// TODO: bad mixing of SDL rendering primitives and app code, here
		sdlPoints := make([]sdl.Point, len(positionReports), len(positionReports))
		var currentPosition sdl.Point
		for i, positionReport := range positionReports {
			realWorldPosition := RealWorldPosition{
				X: positionReport.Lat,
				Y: positionReport.Lon,
			}
			baseMapPosition := view.getBaseMapPosition(realWorldPosition)
			sdlPoints[i] = sdl.Point{
				X: int32(baseMapPosition.X + 0.5),
				Y: int32(baseMapPosition.Y + 0.5),
			}
			currentPosition = sdlPoints[i]
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
		(position.X - view.LRGeo.X) / (view.ULGeo.X - view.LRGeo.X) * view.width,
		(position.Y - view.LRGeo.Y) / (view.ULGeo.Y - view.LRGeo.Y) * view.height,
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
	ULGeo         RealWorldPosition
	LRGeo         RealWorldPosition
	width, height float64
}

type SpriteSheet struct {
	Tex        *sdl.Texture
	SpriteSize int
}

type Sprite struct {
	sheetX, sheetY uint64
}

package main

import (
	"github.com/veandco/go-sdl2/sdl"
	image "github.com/veandco/go-sdl2/sdl_image"
)

const (
	resDir      string = "./res"
	baseMapFile string = "/base.png"
	spriteFile  string = "/sprites.png"
	infoFile    string = "/info"

	spriteSize = 8 // pixels square
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

func InitResources(renderer *sdl.Renderer) error {
	for index := range ViewList {
		viewName := ViewList[index]
		logger.Info("Loading view", "viewName", viewName)

		baseTexture, err := image.LoadTexture(renderer, getResourcePath(ViewList[index], baseMapFile))
		if err != nil {
			return err
		}

		spriteTexture, err := image.LoadTexture(renderer, getResourcePath(ViewList[index], spriteFile))
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
		Views[index] = &View{baseMap, ViewList[index], sprites}
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
	ViewName string
	Sprites  *SpriteSheet
}

// Clears the renderer and redisplays the base map.
// TODO: Redisplay the current tracks as well
func (view *View) Redisplay(renderer *sdl.Renderer) error {
	err := renderer.Clear()
	if err != nil {
		return err
	}

	err = renderer.Copy(view.BaseMap.Tex, nil, nil)
	if err != nil {
		return err
	}

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

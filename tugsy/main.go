package main

import (
	"os"

	"github.com/veandco/go-sdl2/sdl"
	image "github.com/veandco/go-sdl2/sdl_image"
)

const (
	screen_width  int    = 480
	screen_height int    = 600
	screen_title  string = "Tugsy"
)

var updateEvents chan bool = make(chan bool) // Position, etc

func run() int {
	logger.Info("Starting Tugsy")
	window, err := sdl.CreateWindow(
		screen_title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, screen_width, screen_height, sdl.WINDOW_SHOWN)
	if err != nil {
		logger.Fatal("Failed to create window", "err", err)
		return 1
	}
	defer window.Destroy()

	logger.Info("Creating renderer")
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		logger.Fatal("Failed to create renderer", "err", err)
		return 1
	}
	defer renderer.Destroy()

	logger.Info("Initializing image.PNG")
	png_init := image.Init(image.INIT_PNG)
	if png_init != image.INIT_PNG {
		logger.Fatal("Failed to load INIT_PNG", "png_init", png_init)
		return 1
	}
	defer image.Quit()

	logger.Info("Initializing resources")
	err = InitResources(renderer)
	if err != nil {
		logger.Fatal("Could not load resources", "err", err)
		return 1
	}
	defer TeardownResources()

	logger.Info("Initializing the first view")
	view := currentView()
	err = view.Redisplay(renderer)
	if err != nil {
		logger.Fatal("Could not initialize the display with the first view", "err", err)
		return 1
	}
	renderer.Present()

	logger.Info("Starting the UI loop")
	returnCode := 0
	running := true
	for running {
		// PollEvent has to be run in the video init's thread
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyDownEvent:
				switch t.Keysym.Sym {
				case sdl.K_SPACE:
					view := nextView()
					err := view.Redisplay(renderer)
					if err != nil {
						logger.Fatal("Could not rebuild the display", "viewName", view.ViewName, "err", err)
						returnCode = 1
						running = false
					} else {
						renderer.Present()
					}
				}
			}

		}
	}

	return returnCode
}

func masterBlaster() {

}

func main() {
	os.Exit(run())
}

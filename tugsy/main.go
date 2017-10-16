package main

import (
	"os"

	"github.com/andmarios/aislib"
	"github.com/veandco/go-sdl2/sdl"
	image "github.com/veandco/go-sdl2/sdl_image"
)

// STARTUP:
// * Load config
// * Load data
// * Load router
// * Load screen
// * Start routers
// * Master loop
// * UI loop

var TheDisplay *View
var TheData *AISData
var running = true

func run() int {
	logger.Info("Starting Tugsy")
	logger.Info("Loading config")
	config, err := LoadConfig()
	if err != nil {
		logger.Fatal("Could not load the config", "err", err)
		running = false
		return 1
	}

	logger.Info("Loading the AIS routers")
	decoded := make(chan aislib.Message)
	failed := make(chan aislib.FailedSentence)
	routers, err := RemoteAISServersFromConfig(decoded, failed, config)
	if err != nil {
		// TODO: This potentially leaves routers in a dirty state
		logger.Fatal("Could not initialize the routers", "err", err)
		running = false
		return 1
	}

	defer func() {
		for _, r := range routers {
			err := r.stop()
			if err != nil {
				logger.Warn("Error while stopping a router", "sourceName", r.sourceName, "err", err)
			}
		}
	}()

	logger.Info("Creating windows")
	window, err := sdl.CreateWindow(
		screenTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, screenWidth, screenHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		logger.Fatal("Failed to create window", "err", err)
		running = false
		return 1
	}
	defer window.Destroy()

	logger.Info("Creating renderer")
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		logger.Fatal("Failed to create renderer", "err", err)
		running = false
		return 1
	}
	defer renderer.Destroy()

	logger.Info("Initializing image.PNG")
	pngInit := image.Init(image.INIT_PNG)
	if pngInit != image.INIT_PNG {
		logger.Fatal("Failed to load INIT_PNG", "png_init", pngInit)
		running = false
		return 1
	}
	defer image.Quit()

	logger.Info("Initializing resources")
	err = InitResources(renderer)
	if err != nil {
		logger.Fatal("Could not load resources", "err", err)
		running = false
		return 1
	}
	defer TeardownResources()

	logger.Info("Initializing the display")
	TheDisplay = currentView()
	err = TheDisplay.Redisplay(renderer)
	if err != nil {
		logger.Fatal("Could not initialize the display with the first view", "err", err)
		running = false
		return 1
	}
	renderer.Present()

	TheData = NewAISData()
	logger.Info("Starting the position culling loop")
	go TheData.PrunePositions()

	logger.Info("Starting the AIS update loops")
	for _, r := range routers {
		go r.DecodePositions(decoded, failed)
	}

	returnCode := 0

	logger.Info("Starting the UI loop")
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

func main() {
	os.Exit(run())
}

package main

import (
	"os"
	"os/signal"

	"github.com/andmarios/aislib"
	"github.com/joemadeus/tugsy/tugsy/config"
	"github.com/joemadeus/tugsy/tugsy/shipdata"
	"github.com/joemadeus/tugsy/tugsy/views"
	image "github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

// STARTUP:
// * Load config
// * Load data
// * Load router
// * Load screen
// * Start routers
// * Master loop
// * UI loop

const (
	targetFPS uint32 = 30
)

type State struct {
	running bool
}

func run() int {
	logger.Info("Starting Tugsy")
	logger.Info("Loading config")
	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal("Could not load the config", "err", err)
		return 1
	}

	logger.Info("Loading the AIS routers")
	decoded := make(chan aislib.Message)
	failed := make(chan aislib.FailedSentence)
	routers, err := shipdata.RemoteAISServersFromConfig(decoded, failed, cfg)
	if err != nil {
		// TODO: This potentially leaves routers in a dirty state
		logger.Fatal("Could not initialize the routers", "err", err)
		return 1
	}

	defer func() {
		for _, r := range routers {
			r.Stop()
		}
	}()

	logger.Info("Starting the router maintenance loop")
	for _, r := range routers {
		r.Start()
	}

	logger.Info("Starting the AIS update loops")
	for _, r := range routers {
		go r.DecodePositions(decoded, failed)
	}

	logger.Info("Starting the position culling loop")
	go shipdata.PositionData.PrunePositions()

	logger.Info("Initializing image.PNG")
	pngInit := image.Init(image.INIT_PNG)
	if pngInit != image.INIT_PNG {
		logger.Fatal("Failed to load INIT_PNG", "png_init", pngInit)
		return 1
	}
	defer image.Quit()

	logger.Info("Creating windows")
	window, err := sdl.CreateWindow(
		views.ScreenTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, views.ScreenWidth, views.ScreenHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		logger.Fatal("Failed to create window", "err", err)
		return 1
	}
	defer window.Destroy()

	logger.Info("Creating screen renderer")
	screenRenderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		logger.Fatal("Failed to create renderer", "err", err)
		return 1
	}
	defer screenRenderer.Destroy()
	screenRenderer.SetDrawBlendMode(sdl.BLENDMODE_NONE)

	logger.Info("Initializing view resources")

	viewSet, err := views.ViewSetFromConfig(screenRenderer, cfg)
	if err != nil {
		logger.Fatal("Could not load views from config", "err", err)
		return 1
	}
	defer viewSet.TeardownResources()

	logger.Info("Initializing the display")
	currentView := viewSet.CurrentView()
	err = currentView.Display()
	if err != nil {
		logger.Fatal("Could not initialize the display with the first view", "err", err)
		return 1
	}

	returnCode := -1
	var ticks uint32
	var delayMillis uint32
	delayMillis = 1000 / targetFPS

	logger.Info("Adding signal handler")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	logger.Info("Starting the UI loop")
	for returnCode == -1 {
		ticks = sdl.GetTicks()

		// See if we should quit on interrupt
		select {
		case <-signalChan:
			logger.Info("Terminating on interrupt signal")
			returnCode = 0
			continue
		}

		// PollEvent has to be run in the video init's thread
		event := sdl.PollEvent()
		switch t := event.(type) {
		case *sdl.QuitEvent:
			returnCode = 0
		case *sdl.KeyDownEvent:
			switch t.Keysym.Sym {
			case sdl.K_SPACE:
				currentView = viewSet.NextView()
			}
		}

		// Redisplay
		err = currentView.Display()
		if err != nil {
			logger.Fatal("Could not refresh the display", "viewName", currentView.ViewName, "err", err)
			returnCode = 128
			break
		}

		// cap the frame rate to targetFPS
		if ticks < delayMillis {
			sdl.Delay(delayMillis - ticks)
		}
	}

	return returnCode
}

func main() {
	os.Exit(run())
}

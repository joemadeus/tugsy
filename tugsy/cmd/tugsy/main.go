package main

import (
	"os"
	"os/signal"

	"github.com/andmarios/aislib"
	"github.com/joemadeus/tugsy/tugsy/config"
	"github.com/joemadeus/tugsy/tugsy/shipdata"
	"github.com/joemadeus/tugsy/tugsy/views"
	logger "github.com/sirupsen/logrus"
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
	targetFPS uint32 = 15
)

type State struct {
	running bool
}

func run() int {
	logger.Info("Starting Tugsy")
	logger.Info("Loading config")
	cfg, err := config.NewConfig()
	if err != nil {
		logger.WithError(err).Fatal("Could not load the config")
		return 1
	}

	positionData := shipdata.NewAISData()

	logger.Info("Starting the position culling loop")
	go positionData.PrunePositions()

	logger.Info("Loading the AIS routers")
	decoded := make(chan aislib.Message)
	failed := make(chan aislib.FailedSentence)
	routers, err := shipdata.RemoteAISServersFromConfig(decoded, failed, cfg)
	if err != nil {
		logger.WithError(err).Fatal("Could not initialize the routers")
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

	logger.Info("Initializing INIT_PNG")
	pngInit := image.Init(image.INIT_PNG)
	if pngInit != image.INIT_PNG {
		logger.Fatalf("failed to load INIT_PNG, instead got %v", pngInit)
		return 1
	}
	defer image.Quit()

	logger.Info("Creating windows")
	window, err := sdl.CreateWindow(
		views.ScreenTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, views.ScreenWidth, views.ScreenHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		logger.WithError(err).Fatal("failed to create window")
		return 1
	}
	defer window.Destroy()

	logger.Info("Creating screen renderer")
	screenRenderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		logger.WithError(err).Fatal("failed to create renderer")
		return 1
	}
	defer screenRenderer.Destroy()
	screenRenderer.SetDrawBlendMode(sdl.BLENDMODE_NONE)

	logger.Info("Initializing view resources")
	spriteSet, err := views.NewSpriteSet(screenRenderer, cfg)
	if err != nil {
		logger.WithError(err).Fatal("could not load sprites from config")
		return 1
	}

	infoPane, err := views.NewInfoElement(screenRenderer, cfg)
	if err != nil {
		logger.WithError(err).Fatal("could not load the base info pane")
		return 1
	}

	renderSet := []views.ViewElement{
		shipdata.NewShipPositionElement(positionData, spriteSet),
		infoPane,
	}

	viewSet, err := views.ViewSetFromConfig(screenRenderer, renderSet, cfg)
	if err != nil {
		logger.WithError(err).Fatal("Could not load views from config")
		return 1
	}
	defer viewSet.Teardown()

	logger.Info("Initializing the display")
	currentView := viewSet.CurrentView()
	if err = currentView.Display(); err != nil {
		logger.WithError(err).Fatal("could not initialize the display with the first view")
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
		if sig := <-signalChan; sig != nil {
			logger.Info("Terminating on interrupt signal")
			returnCode = 0
			continue
		}

		// PollEvent has to be run in the video init's thread
		event := sdl.PollEvent()
		switch t := event.(type) {
		case *sdl.QuitEvent:
			returnCode = 0
			continue
		case *sdl.KeyboardEvent:
			switch {
			case t.Keysym.Sym == sdl.K_SPACE && t.Type == sdl.KEYDOWN:
				currentView = viewSet.NextView()
			}
		}

		// Redisplay
		if err = currentView.Display(); err != nil {
			logger.WithError(err).Fatalf("Could not refresh the display, view %s", currentView.ViewName)
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

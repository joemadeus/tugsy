package main

import (
	"os"

	"github.com/andmarios/aislib"
	"github.com/joemadeus/tugsy/tugsy/config"
	"github.com/joemadeus/tugsy/tugsy/shipdata"
	"github.com/joemadeus/tugsy/tugsy/views"
	logger "github.com/sirupsen/logrus"
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

func run() int {
	logger.Info("Starting Tugsy")
	logger.Info("Loading config")
	cfg, err := config.NewConfig()
	if err != nil {
		logger.WithError(err).Fatal("Could not load the config")
		return 1
	}

	loglevel, err := logger.ParseLevel(cfg.GetString("loglevel"))
	if err != nil {
		logger.WithError(err).Fatalf("bad log level in config")
	}
	logger.SetLevel(loglevel)

	aisData := shipdata.NewAISData()

	logger.Info("Starting the position culling loop")
	go aisData.PrunePositions()

	logger.Info("Loading the AIS routers")
	decoded := make(chan aislib.Message)
	failed := make(chan aislib.FailedSentence)
	routers, err := shipdata.RemoteAISServersFromConfig(aisData, decoded, failed, cfg)
	if err != nil {
		logger.WithError(err).Fatal("Could not initialize the routers")
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

	logger.Info("Initializing INIT_EVERYTHING")
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		logger.WithError(err).Fatalf("failed to INIT_EVERYTHING")
	}
	defer sdl.Quit()

	logger.Info("Creating windows")
	window, err := sdl.CreateWindow(
		views.ScreenTitle,
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		views.ScreenWidth,
		views.ScreenHeight,
		sdl.WINDOW_SHOWN|sdl.WINDOW_OPENGL) // |sdl.WINDOW_BORDERLESS
	if err != nil {
		logger.WithError(err).Fatal("failed to create window")
	}
	defer window.Destroy()

	logger.Info("Creating screen renderer")
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		logger.WithError(err).Fatal("failed to create renderer")
	}
	defer renderer.Destroy()

	if err := renderer.SetDrawBlendMode(sdl.BLENDMODE_NONE); err != nil {
		logger.WithError(err).Fatal("failed to set blend mode")
	}

	logger.Info("Initializing view resources & elements")
	spriteSet, err := views.NewSpriteSet(renderer, cfg)
	if err != nil {
		logger.WithError(err).Fatal("could not load sprites from config")
	}
	logger.Info("Initialized sprites")

	// create the root view element and its direct children
	baseInfoElement, err := views.NewBaseInfoElement(cfg, renderer)
	if err != nil {
		logger.WithError(err).Fatal("Could not initialize BaseInfoElement")
	}
	allPositionsElement := views.NewAllPositionElements(spriteSet, aisData, baseInfoElement)
	rootElement := views.NewRootElement(cfg, baseInfoElement, allPositionsElement)
	logger.Info("Initialized RootElement & children")

	viewSet, err := views.ViewSetFromConfig(cfg, renderer, rootElement)
	if err != nil {
		logger.WithError(err).Fatal("Could not load views from config")
	}
	defer viewSet.Teardown()
	logger.Info("Initialized ViewSet")

	logger.Info("Showing window")
	window.Show()

	logger.Info("Initializing the display")
	currentView := viewSet.CurrentView()
	if err = currentView.Display(); err != nil {
		logger.WithError(err).Fatal("could not initialize the display with the first view")
	}

	returnCode := -1
	delayMillis := 1000 / targetFPS
	var ticks uint32

	logger.Info("Starting the UI loop")
	for returnCode == -1 {
		ticks = sdl.GetTicks()

		// PollEvent has to be run in the video init's thread
		event := sdl.PollEvent()
		switch t := event.(type) {
		case *sdl.QuitEvent:
			returnCode = 0
			continue

		case *sdl.MouseButtonEvent:
			logger.Debugf("MOUSE EVENT: %+v", event)
			if t.Type == sdl.MOUSEBUTTONUP {
				mouseEvent := event.(*sdl.MouseButtonEvent)
				if err := rootElement.Touch(mouseEvent.X, mouseEvent.Y); err != nil {
					logger.WithError(err).Error("unable to handle touch event")
				}
			}

		case *sdl.KeyboardEvent:
			// TODO this should be input from the single button on the front panel
			switch {
			case t.Keysym.Sym == sdl.K_SPACE && t.Type == sdl.KEYDOWN:
				currentView = viewSet.NextView()
			}
		}

		// Redisplay
		if err = currentView.Display(); err != nil {
			logger.WithError(err).Errorf("Could not refresh the display, view %s", currentView.Name)
			returnCode = 128
			continue
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

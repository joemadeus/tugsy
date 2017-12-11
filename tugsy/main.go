package main

import (
	"os"

	"os/signal"

	"github.com/andmarios/aislib"
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

const appConfig = "./config"

var MachineAndProcessState State

type State struct {
	TheDisplay *View
	TheData    *AISData
	running    bool
}

func run() int {
	logger.Info("Starting Tugsy")
	logger.Info("Loading config")
	config, err := LoadConfig(appConfig)
	if err != nil {
		logger.Fatal("Could not load the config", "err", err)
		MachineAndProcessState.running = false
		return 1
	}

	logger.Info("Loading the AIS routers")
	decoded := make(chan aislib.Message)
	failed := make(chan aislib.FailedSentence)
	routers, err := RemoteAISServersFromConfig(decoded, failed, config)
	if err != nil {
		// TODO: This potentially leaves routers in a dirty state
		logger.Fatal("Could not initialize the routers", "err", err)
		MachineAndProcessState.running = false
		return 1
	}

	defer func() {
		for _, r := range routers {
			err := r.stop()
			if err != nil {
				logger.Warn("Error while stopping a router", "sourceName", r.SourceName, "err", err)
			}
		}
	}()

	MachineAndProcessState.TheData = NewAISData()
	logger.Info("Starting the position culling loop")
	go MachineAndProcessState.TheData.PrunePositions()

	logger.Info("Starting the AIS update loops")
	for _, r := range routers {
		go r.DecodePositions(decoded, failed)
	}

	logger.Info("Initializing image.PNG")
	pngInit := image.Init(image.INIT_PNG)
	if pngInit != image.INIT_PNG {
		logger.Fatal("Failed to load INIT_PNG", "png_init", pngInit)
		MachineAndProcessState.running = false
		return 1
	}
	defer image.Quit()

	logger.Info("Creating windows")
	window, err := sdl.CreateWindow(
		screenTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, screenWidth, screenHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		logger.Fatal("Failed to create window", "err", err)
		MachineAndProcessState.running = false
		return 1
	}
	defer window.Destroy()

	logger.Info("Creating screen renderer")
	screenRenderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		logger.Fatal("Failed to create renderer", "err", err)
		MachineAndProcessState.running = false
		return 1
	}
	defer screenRenderer.Destroy()

	logger.Info("Initializing resources")
	err = InitResources(screenRenderer)
	if err != nil {
		logger.Error("Could not load resources", "err", err)
		MachineAndProcessState.running = false
		return 1
	}
	defer TeardownResources()

	logger.Info("Initializing the display")
	MachineAndProcessState.TheDisplay = currentView()
	err = MachineAndProcessState.TheDisplay.Display()
	if err != nil {
		logger.Fatal("Could not initialize the display with the first view", "err", err)
		MachineAndProcessState.running = false
		return 1
	}

	returnCode := 0
	var ticks uint32
	var delayMillis uint32
	delayMillis = 1000 / targetFPS

	logger.Info("Adding signal handler")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	logger.Info("Starting the UI loop")
	MachineAndProcessState.running = true
	// set this to dirty so we display the initial view
	MachineAndProcessState.TheData.dirty = true
	for MachineAndProcessState.running {
		ticks = sdl.GetTicks()

		// See if we should quit on interrupt
		select {
		case <-signalChan:
			logger.Info("Terminating on interrupt signal")
			MachineAndProcessState.running = false
			continue
		default:
		}

		// PollEvent has to be run in the video init's thread
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				MachineAndProcessState.running = false
			case *sdl.KeyDownEvent:
				switch t.Keysym.Sym {
				case sdl.K_SPACE:
					MachineAndProcessState.TheDisplay = nextView()
				}
			}
		}

		// Redisplay
		err = MachineAndProcessState.TheDisplay.Display()
		if err != nil {
			logger.Fatal("Could not refresh the display", "viewName", MachineAndProcessState.TheDisplay.ViewName, "err", err)
			returnCode = 1
			MachineAndProcessState.running = false
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

package main

import (
	"os"
	"github.com/veandco/go-sdl2/sdl"
	image "github.com/veandco/go-sdl2/sdl_image"
	"fmt"
)

const (
	screen_x      int = 0
	screen_y      int = 0
	screen_width  int = 480
	screen_height int = 600
	screen_title  string = "Tugsy"
)

func run() int {
	fmt.Fprint(os.Stdout, "Starting Tugsy\n")
	window, err := sdl.CreateWindow(
		screen_title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, screen_width, screen_height, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer window.Destroy()

	fmt.Fprint(os.Stdout, "Creating renderer\n")
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		return 1
	}
	defer renderer.Destroy()

	fmt.Fprint(os.Stdout, "Initializing image.PNG\n")
	png_init := image.Init(image.INIT_PNG)
	if png_init != image.INIT_PNG {
		fmt.Fprintf(os.Stderr, "Failed to load INIT_PNG: %s\n", png_init)
		return 1
	}
	defer image.Quit()

	fmt.Fprint(os.Stdout, "Initializing resources\n")
	err = InitResources(renderer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not load resources: %s\n", err)
		return 1
	}
	defer TeardownResources()

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyDownEvent:
				switch t.Keysym.Sym {
				case sdl.K_SPACE:
					view := nextView()
					err := view.redisplay(renderer)
					if err != nil { fmt.Fprint(os.Stderr, "Could not rebuild the display for %s: %s\n", view.ViewName, err) }
				}
			}
		}
	}

	return 0
}

func main() {
	os.Exit(run())
}

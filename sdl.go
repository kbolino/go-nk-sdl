package nksdl

import (
	"errors"
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

// SDLDriver is implemented by any type capable of initializing SDL and creating
// its core resources.
type SDLDriver interface {
	InitSDL() error
	CreateWindow() (*sdl.Window, error)
	CreateRenderer(window *sdl.Window) (*sdl.Renderer, error)
}

// DefaultSDLDriver is the default implementation of SDLDriver. It can be used
// directly as an SDLDriver, or extended to patch/override its behavior.
type DefaultSDLDriver struct {
	// InitFlags contains flags to pass to sdl.Init.
	InitFlags uint32
	// Hints contains hint keys and values to pass to sdl.SetHint.
	Hints map[string]string
	// Window contains options for creating the window.
	Window WindowOpts
	// Render contains options for creating the renderer.
	Render RenderOpts
}

var _ SDLDriver = &DefaultSDLDriver{}

func (d *DefaultSDLDriver) InitSDL() error {
	if err := sdl.Init(d.InitFlags); err != nil {
		return err
	}
	for key, value := range d.Hints {
		sdl.SetHint(key, value)
	}
	return nil
}

func (d *DefaultSDLDriver) CreateWindow() (*sdl.Window, error) {
	window, err := sdl.CreateWindow(d.Window.Title, d.Window.PosX, d.Window.PosY, d.Window.Width, d.Window.Height,
		d.Window.Flags)
	if err != nil {
		return nil, err
	}
	return window, err
}

func (d *DefaultSDLDriver) CreateRenderer(window *sdl.Window) (*sdl.Renderer, error) {
	renderDriver := -1
	if len(d.Render.Drivers) != 0 {
		numRenderDrivers, err := sdl.GetNumRenderDrivers()
		if err != nil {
			return nil, fmt.Errorf("getting number of render drivers: %w", err)
		}
		infos := make([]sdl.RendererInfo, numRenderDrivers)
		for i := 0; i < numRenderDrivers; i++ {
			if _, err := sdl.GetRenderDriverInfo(i, &infos[i]); err != nil {
				return nil, fmt.Errorf("getting info for render driver %d: %w", i, err)
			}
		}
	preferenceLoop:
		for _, driver := range d.Render.Drivers {
			for i := range infos {
				if infos[i].Name == driver {
					renderDriver = i
					break preferenceLoop
				}
			}
		}
		if renderDriver < 0 {
			return nil, errors.New("could not find any preferred render driver")
		}
	}
	renderer, err := sdl.CreateRenderer(window, renderDriver, d.Render.Flags)
	if err != nil {
		return nil, fmt.Errorf("creating renderer: %w", err)
	}
	return renderer, nil
}

// RenderOpts sets options for DefaultSDLDriver.CreateRenderer.
type RenderOpts struct {
	Drivers []string
	Flags   uint32
}

// WindowOpts sets options for DefaultSDLDriver.CreateWindow.
type WindowOpts struct {
	Title         string
	PosX, PosY    int32
	Width, Height int32
	Flags         uint32
}

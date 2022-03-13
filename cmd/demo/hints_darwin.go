package main

import "github.com/veandco/go-sdl2/sdl"

func addSDLHints() {
	// TODO: figure out why window borders and titles don't render in Metal
	sdl.SetHint(sdl.HINT_RENDER_DRIVER, "opengl")
}

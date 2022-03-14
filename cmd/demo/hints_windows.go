package main

import "github.com/veandco/go-sdl2/sdl"

func addSDLHints() {
	sdl.SetHint(sdl.HINT_RENDER_DRIVER, "direct3d11")
}

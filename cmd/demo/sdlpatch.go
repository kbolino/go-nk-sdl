package main

// #cgo LDFLAGS: -lSDL2
// #include "SDL2/SDL.h"
import "C"

import (
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

func RenderGeometryRaw(
	renderer *sdl.Renderer,
	texture *sdl.Texture,
	xy unsafe.Pointer, xyStride int32,
	color unsafe.Pointer, colorStride int32,
	uv unsafe.Pointer, uvStride int32,
	numVertices int32,
	indices unsafe.Pointer, numIndices, sizeIndices int32,
) error {
	// int SDL_RenderGeometryRaw(
	//     SDL_Renderer *renderer,
	//     SDL_Texture *texture,
	//     const float *xy, int xy_stride,
	//     const SDL_Color *color, int color_stride,
	//     const float *uv, int uv_stride,
	//     int num_vertices,
	//     const void *indices, int num_indices, int size_indices
	// );
	result := C.SDL_RenderGeometryRaw(
		(*C.SDL_Renderer)(renderer),
		(*C.SDL_Texture)(texture),
		(*C.float)(xy), C.int(xyStride),
		(*C.SDL_Color)(color), C.int(colorStride),
		(*C.float)(uv), C.int(uvStride),
		C.int(numVertices),
		indices, C.int(numIndices), C.int(sizeIndices),
	)
	if result != 0 {
		return sdl.GetError()
	}
	return nil
}

package nksdl

import (
	"reflect"
	"unsafe"

	"github.com/kbolino/go-nk"
	"github.com/veandco/go-sdl2/sdl"
)

// N.B. can't use a generic handle-to-pointer function until the following issue
// in Go is resolved:
// https://github.com/golang/go/issues/51733 (Go 1.19)
// https://github.com/golang/go/issues/51741 (Go 1.18.1)

func textureToHandle(tex *sdl.Texture) nk.Handle {
	return nk.Handle(unsafe.Pointer(tex))
}

func handleToTexture(handle nk.Handle) *sdl.Texture {
	return (*sdl.Texture)(unsafe.Pointer(handle))
}

func reinterpretSlice[T any](p []byte, size int) []T {
	// adapted from https://stackoverflow.com/a/11927363/814422
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&p))
	header.Len /= size
	header.Cap /= size
	return *(*[]T)(unsafe.Pointer(&header))
}

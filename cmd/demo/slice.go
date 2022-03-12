package main

import (
	"reflect"
	"unsafe"
)

func reinterpretSlice[T any](p []byte, size int) []T {
	// adapted from https://stackoverflow.com/a/11927363/814422
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&p))
	header.Len /= size
	header.Cap /= size
	return *(*[]T)(unsafe.Pointer(&header))
}

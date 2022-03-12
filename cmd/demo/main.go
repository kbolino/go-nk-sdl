package main

import (
	"fmt"
	"os"
	"runtime"
	"unsafe"

	"github.com/kbolino/go-nk"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	sizeofVertex           = 20 // sizeof(sdl.Vertex)
	offsetofVertexPosition = 0  // offsetof(sdl.Vertex, Position)
	offsetofVertexColor    = 8  // offsetof(sdl.Vertex, Color)
	offsetofVertexTexCoord = 12 // offsetof(sdl.Vertex, TexCoord)

	segmentCount = 22 // magic number?
	aa           = nk.AntiAliasingOn
)

var (
	window     *sdl.Window
	renderer   *sdl.Renderer
	sdlFontTex *sdl.Texture

	nkc    *nk.Context
	config *nk.ConvertConfig

	cbuf, ebuf, vbuf *nk.Buffer
)

func init() {
	runtime.LockOSThread()
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}

func run() (err error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return fmt.Errorf("initializing SDL: %w", err)
	}
	defer sdl.Quit()

	sdl.SetHint(sdl.HINT_VIDEO_HIGHDPI_DISABLED, "0")

	window, err = sdl.CreateWindow("go-nk-sdl demo", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, 800, 600,
		sdl.WINDOW_ALLOW_HIGHDPI)
	if err != nil {
		return fmt.Errorf("creating window: %w", err)
	}
	defer func() {
		if err2 := window.Destroy(); err2 != nil && err == nil {
			err = fmt.Errorf("destroying window: %w", err2)
		}
	}()

	renderer, err = sdl.CreateRenderer(window, -1, 0)
	if err != nil {
		return fmt.Errorf("creating renderer: %w", err)
	}
	defer func() {
		if err2 := renderer.Destroy(); err2 != nil && err == nil {
			err = fmt.Errorf("destroying renderer: %w", err2)
		}
	}()

	renderScale := int32(1)
	if renderW, renderH, err := renderer.GetOutputSize(); err != nil {
		return fmt.Errorf("getting renderer output size: %w", err)
	} else {
		windowW, windowH := window.GetSize()
		horizScale := renderW / windowW
		vertScale := renderH / windowH
		if horizScale != vertScale {
			return fmt.Errorf("render output is not scaled uniformly: renderer (%d x %d), window (%d x %d)",
				renderW, renderH, windowW, windowH)
		}
		renderScale = horizScale
	}
	renderScaleF := float32(renderScale)

	if nkc, err = nk.NewContext(); err != nil {
		return fmt.Errorf("initializing nuklear: %w", err)
	}
	defer nkc.Free()

	nkfa := nk.NewFontAtlas()
	defer nkfa.Free()
	nkfa.Begin()
	defaultFont := nkfa.AddDefaultFont(14)
	image, width, height := nkfa.Bake(nk.FontAtlasRGBA32)
	fmt.Println("baked image width =", width, "height =", height)
	sdlFontTex, err = renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STATIC, width, height)
	if err != nil {
		return fmt.Errorf("creating font texture; %w", err)
	}
	if err = sdlFontTex.Update(nil, image, int(4*width)); err != nil {
		return fmt.Errorf("uploading font atlas to texture: %w", err)
	}
	if err = sdlFontTex.SetBlendMode(sdl.BLENDMODE_BLEND); err != nil {
		return fmt.Errorf("setting texture blend mode: %w", err)
	}
	drawNullTex := nkfa.End(nk.Handle(unsafe.Pointer(sdlFontTex)))
	nkc.StyleSetFont(defaultFont.Handle())
	//nkfa.Cleanup()

	configBuilder := nk.ConvertConfigBuilder{
		CConvertConfig: nk.CConvertConfig{
			VertexSize:         sizeofVertex,
			VertexAlignment:    4, // alignof(sdl.Vertex)
			Null:               drawNullTex,
			CircleSegmentCount: segmentCount,
			CurveSegmentCount:  segmentCount,
			ArcSegmentCount:    segmentCount,
			GlobalAlpha:        1.0,
			ShapeAA:            aa,
			LineAA:             aa,
		},
		VertexLayout: []nk.DrawVertexLayoutElement{
			{Attribute: nk.VertexPosition, Format: nk.FormatFloat, Offset: offsetofVertexPosition},
			{Attribute: nk.VertexColor, Format: nk.FormatR8G8B8A8, Offset: offsetofVertexColor},
			{Attribute: nk.VertexTexcoord, Format: nk.FormatFloat, Offset: offsetofVertexTexCoord},
		},
	}
	config = configBuilder.Build()
	defer config.Free()

	cbuf = nk.NewBuffer()
	defer cbuf.Free()
	ebuf = nk.NewBuffer()
	defer ebuf.Free()
	vbuf = nk.NewBuffer()
	defer vbuf.Free()

	checked := false
outer:
	for {
		nkc.InputBegin()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			if quit := convertEvent(nkc, event, renderScale); quit {
				break outer
			}
		}
		nkc.InputEnd()

		// if (nk_begin(ctx, "Demo", nk_rect(50, 50, 230, 250),
		//     NK_WINDOW_BORDER|NK_WINDOW_MOVABLE|NK_WINDOW_SCALABLE|
		//     NK_WINDOW_MINIMIZABLE|NK_WINDOW_TITLE))
		nkc.Begin("Demo",
			&nk.Rect{X: renderScaleF * 50, Y: renderScaleF * 50,
				W: renderScaleF * 230, H: renderScaleF * 250},
			nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowMinimizable|nk.WindowTitle,
		)
		nkc.LayoutRowStatic(renderScaleF*30, renderScale*81, 1)
		if nkc.ButtonText("Button") {
			fmt.Println("button pressed")
		}
		nkc.LayoutRowStatic(renderScaleF*30, renderScale*80, 1)
		checked = nkc.CheckText("Check me", checked)
		nkc.End()

		if err = renderer.SetDrawColor(25, 45, 61, 255); err != nil {
			return fmt.Errorf("setting renderer draw color: %w", err)
		}
		if err = renderer.Clear(); err != nil {
			return fmt.Errorf("clearing renderer: %w", err)
		}

		cbuf.Clear()
		ebuf.Clear()
		vbuf.Clear()
		if err := nkc.Convert(cbuf, vbuf, ebuf, config); err != nil {
			return fmt.Errorf("convert error: %w", err)
		}
		ebufMem := ebuf.Memory()
		// technically nk_draw_index is uint32 but it's unlikely to ever use
		// the high bit
		indices := reinterpretSlice[int32](ebufMem, 4)
		oldClipRect := renderer.GetClipRect()
		var err error
		nkc.DrawForEach(cbuf, func(cmd *nk.DrawCommand) (ok bool) {
			if cmd.ElemCount == 0 {
				return true
			}
			rect := sdl.Rect{
				X: int32(cmd.ClipRect.X),
				Y: int32(cmd.ClipRect.Y),
				W: int32(cmd.ClipRect.W),
				H: int32(cmd.ClipRect.H),
			}
			if err = renderer.SetClipRect(&rect); err != nil {
				err = fmt.Errorf("setting renderer clip rectangle: %w", err)
				return false
			}

			vbufMem := vbuf.Memory()
			vertices := reinterpretSlice[sdl.Vertex](vbufMem, sizeofVertex)
			err = renderer.RenderGeometry(
				(*sdl.Texture)(unsafe.Pointer(cmd.Texture)),
				vertices,
				indices[:cmd.ElemCount],
			)
			// vbufVertices := vbufMem[byteOffsetVertex:]
			// vbufColors := vbufMem[byteOffsetColor:]
			// vbufUVs := vbufMem[byteOffsetUV:]
			// err = RenderGeometryRaw(
			// 	renderer,
			// 	(*sdl.Texture)(unsafe.Pointer(cmd.Texture)),
			// 	unsafe.Pointer(&vbufVertices[0]), byteSizeVertex,
			// 	unsafe.Pointer(&vbufColors[0]), byteSizeVertex,
			// 	unsafe.Pointer(&vbufUVs[0]), byteSizeVertex,
			// 	int32(numVertices),
			// 	unsafe.Pointer(&ebufMem[ebufOffset]), int32(cmd.ElemCount), 2,
			// )
			if err != nil {
				err = fmt.Errorf("rendering raw geometry: %w", err)
				return false
			}
			indices = indices[cmd.ElemCount:]
			return true
		})
		if err != nil {
			return fmt.Errorf("executing draw commands: %w", err)
		}
		if oldClipRect.W != 0 && oldClipRect.H != 0 {
			if err := renderer.SetClipRect(&oldClipRect); err != nil {
				return fmt.Errorf("restoring renderer clip rectangle: %w", err)
			}
		}

		renderer.Present()
		nkc.Clear()
		cbuf.Clear()
		ebuf.Clear()
		vbuf.Clear()
	}
	return nil
}

func convertEvent(nkc *nk.Context, event sdl.Event, scale int32) (quit bool) {
	window.GetWMInfo()
	switch e := event.(type) {
	case *sdl.QuitEvent:
		return true
	case *sdl.MouseMotionEvent:
		nkc.InputMotion(scale*e.X, scale*e.Y)
	case *sdl.MouseButtonEvent:
		x, y := scale*e.X, scale*e.Y
		down := false
		if e.State == sdl.PRESSED {
			down = true
		}
		switch e.Button {
		case sdl.BUTTON_LEFT:
			if e.Clicks == 2 {
				nkc.InputButton(nk.ButtonDouble, x, y, down)
			}
			nkc.InputButton(nk.ButtonLeft, x, y, down)
		case sdl.BUTTON_RIGHT:
			nkc.InputButton(nk.ButtonRight, x, y, down)
		case sdl.BUTTON_MIDDLE:
			nkc.InputButton(nk.ButtonMiddle, x, y, down)
		}
	case *sdl.MouseWheelEvent:
		// TODO scale scroll or no?
		nkc.InputScroll(e.PreciseX, e.PreciseY)
	case *sdl.KeyboardEvent:
		var down bool
		if e.State == sdl.PRESSED {
			down = true
		}
		var shift bool
		if e.Keysym.Mod&sdl.KMOD_SHIFT != 0 {
			shift = true
		}
		var ctrl bool
		if e.Keysym.Mod&sdl.KMOD_CTRL != 0 {
			ctrl = true
		}
		switch e.Keysym.Sym {
		case sdl.K_LSHIFT, sdl.K_RSHIFT:
			nkc.InputKey(nk.KeyShift, down)
		case sdl.K_DELETE:
			nkc.InputKey(nk.KeyDel, down)
		case sdl.K_RETURN:
			nkc.InputKey(nk.KeyEnter, down)
		case sdl.K_TAB:
			nkc.InputKey(nk.KeyTab, down)
		case sdl.K_BACKSPACE:
			nkc.InputKey(nk.KeyBackspace, down)
		case sdl.K_HOME:
			nkc.InputKey(nk.KeyTextStart, down)
			nkc.InputKey(nk.KeyScrollStart, down)
		case sdl.K_END:
			nkc.InputKey(nk.KeyTextEnd, down)
			nkc.InputKey(nk.KeyScrollEnd, down)
		case sdl.K_PAGEUP:
			nkc.InputKey(nk.KeyScrollUp, down)
		case sdl.K_PAGEDOWN:
			nkc.InputKey(nk.KeyScrollDown, down)
		case sdl.K_z:
			nkc.InputKey(nk.KeyTextUndo, down && ctrl && !shift)
			nkc.InputKey(nk.KeyTextRedo, down && ctrl && shift)
		case sdl.K_c:
			nkc.InputKey(nk.KeyCopy, down && ctrl)
		case sdl.K_v:
			nkc.InputKey(nk.KeyPaste, down && ctrl)
		case sdl.K_x:
			nkc.InputKey(nk.KeyCut, down && ctrl)
		case sdl.K_a:
			nkc.InputKey(nk.KeyTextLineStart, down && ctrl)
		case sdl.K_e:
			nkc.InputKey(nk.KeyTextLineEnd, down && ctrl)
		case sdl.K_UP:
			nkc.InputKey(nk.KeyUp, down)
		case sdl.K_DOWN:
			nkc.InputKey(nk.KeyDown, down)
		case sdl.K_LEFT:
			nkc.InputKey(nk.KeyLeft, down)
		case sdl.K_RIGHT:
			nkc.InputKey(nk.KeyRight, down)
		}
	case *sdl.TextInputEvent:
		for _, r := range e.GetText() {
			nkc.InputUnicode(r)
		}
	}
	return false
}

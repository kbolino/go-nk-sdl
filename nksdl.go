// Package nksdl provides integration between Nuklear and SDL2 via the
// github.com/kbolino/go-nk and github.com/veandco/go-sdl2 modules.
package nksdl

import (
	"errors"
	"fmt"

	"github.com/kbolino/go-nk"
	"github.com/veandco/go-sdl2/sdl"
)

// Driver is the main focus of NkSDL. It is used to manage SDL resources and
// Nuklear resources through initial setup, per-frame rendering, and pre-quit
// teardown. Driver delegates SDL-related setup to SDLDriver and Nuklear-related
// setup to NkDriver.
type Driver struct {
	sdlDriver     SDLDriver
	nkDriver      NkDriver
	eventListener EventListener
	eventHandler  EventHandler

	window   *sdl.Window
	renderer *sdl.Renderer
	fontTex  *sdl.Texture

	context     *nk.Context
	atlas       *nk.FontAtlas
	font        *nk.Font
	largeFont   *nk.Font
	null        nk.DrawNullTexture
	convertConf *nk.ConvertConfig
	commands    *nk.Buffer
	elements    *nk.Buffer
	vertices    *nk.Buffer

	uiScale       float32
	bgColor       sdl.Color
	clampClipRect bool
}

// NewDriver creates a new Driver from the given parameters. The sdlDriver and
// nkDriver must not be nil, or else NewDriver will panic. The bindings map is
// used to map keys to Nuklear actions. The eventListener is optional, but if
// non-nil will be called for every SDL event after it is handled by Nuklear.
func NewDriver(
	sdlDriver SDLDriver,
	nkDriver NkDriver,
	bindings map[KeyInput]KeyAction,
	eventListener EventListener,
) *Driver {
	if sdlDriver == nil {
		panic("sdlDriver is nil")
	} else if nkDriver == nil {
		panic("nkDriver is nil")
	}
	return &Driver{
		sdlDriver:     sdlDriver,
		nkDriver:      nkDriver,
		eventListener: eventListener,
		eventHandler:  NewEventHandler(bindings),
		uiScale:       1,
	}
}

func (d *Driver) Window() *sdl.Window {
	return d.window
}

func (d *Driver) Renderer() *sdl.Renderer {
	return d.renderer
}

func (d *Driver) Context() *nk.Context {
	return d.context
}

func (d *Driver) SetBGColor(color sdl.Color) {
	d.bgColor = color
}

func (d *Driver) SetUIScale(uiScale float32) error {
	if uiScale != uiScale || uiScale < 0 || uiScale > 5 {
		return fmt.Errorf("uiScale(%g) is out of bounds", uiScale)
	}
	if uiScale == 0 {
		if err := d.computeUIScale(); err != nil {
			return fmt.Errorf("computing UI scale: %w", err)
		}
	} else {
		d.uiScale = uiScale
	}
	return nil
}

// Init initializes the Driver, creating the SDL window and renderer as well
// as the Nuklear context and fonts. Init should be called once in the lifetime
// of a Driver, before any calls to PreRender.
func (d *Driver) Init() error {
	var err error
	defer func() {
		if err != nil {
			d.Destroy()
		}
	}()
	if err = d.sdlDriver.InitSDL(); err != nil {
		return fmt.Errorf("initializing SDL: %w", err)
	}
	if d.window, err = d.sdlDriver.CreateWindow(); err != nil {
		return fmt.Errorf("creating SDL window: %w", err)
	}
	if d.renderer, err = d.sdlDriver.CreateRenderer(d.window); err != nil {
		return fmt.Errorf("creating SDL renderer: %w", err)
	}
	if info, err := d.renderer.GetInfo(); err != nil {
		return fmt.Errorf("getting SDL renderer info: %w", err)
	} else if info.Name == "metal" {
		var ver sdl.Version
		sdl.GetVersion(&ver)
		if sdl.VERSIONNUM(int(ver.Major), int(ver.Minor), int(ver.Patch)) < sdl.VERSIONNUM(2, 0, 22) {
			// fixes https://discourse.libsdl.org/t/rendergeometryraw-producing-different-results-in-metal-vs-opengl/34953
			d.clampClipRect = true
		}
	}
	if d.context, err = d.nkDriver.CreateContext(); err != nil {
		return fmt.Errorf("creating Nuklear context: %w", err)
	}
	if d.atlas, err = d.nkDriver.CreateFontAtlas(); err != nil {
		return fmt.Errorf("creating font atlast: %w", err)
	}
	if d.font, err = d.nkDriver.CreateFont(d.atlas, 1); err != nil {
		return fmt.Errorf("creating font: %w", err)
	}
	if d.largeFont, err = d.nkDriver.CreateFont(d.atlas, 2); err != nil {
		return fmt.Errorf("creating large font: %w", err)
	}
	if d.null, err = d.bakeFont(); err != nil {
		return fmt.Errorf("baking font: %w", err)
	}
	d.largeFont.ScaleHeight(2)
	d.convertConf = d.nkDriver.CreateConvertConfig(
		vertexLayout,
		uint32(vertexSize),
		uint32(vertexAlignment),
		d.null,
	)
	d.commands = nk.NewBuffer()
	d.elements = nk.NewBuffer()
	d.vertices = nk.NewBuffer()
	return nil
}

// PreRender preforms early render phase actions, including polling for events,
// mapping input events to actions, and clearing the renderer. PreRender should
// be called once at the beginning of every frame.
func (d *Driver) PreRender() (err error) {
	d.context.Clear()
	d.context.InputBegin()
	defer d.context.InputEnd()
	alive := true
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		eventType, usedByNuklear := d.eventHandler.HandleEvent(d.context, event)
		if d.eventListener != nil {
			if err = d.eventListener(event, eventType, usedByNuklear); err == ErrQuit {
				alive = false
			} else if err != nil {
				return fmt.Errorf("passing event %#v to event listener: %w", event, err)
			}
		} else if eventType == EventTypeQuit {
			alive = false
		}
	}
	d.context.InputEnd()
	if !alive {
		return ErrQuit
	}
	var oldR, oldG, oldB, oldA uint8
	if oldR, oldG, oldB, oldA, err = d.renderer.GetDrawColor(); err != nil {
		return fmt.Errorf("getting renderer draw color: %w", err)
	}
	defer func() {
		if err2 := d.renderer.SetDrawColor(oldR, oldG, oldB, oldA); err2 != nil && err == nil {
			err = fmt.Errorf("restoring renderer draw color: %w", err)
		}
	}()
	if err = d.renderer.SetDrawColor(d.bgColor.R, d.bgColor.G, d.bgColor.B, d.bgColor.A); err != nil {
		return fmt.Errorf("setting renderer draw color: %w", err)
	}
	if err = d.renderer.Clear(); err != nil {
		return fmt.Errorf("clearing renderer: %w", err)
	}
	if d.uiScale > 1.5 {
		d.context.StyleSetFont(d.largeFont.Handle())
	} else {
		d.context.StyleSetFont(d.font.Handle())
	}
	return nil
}

// PostRender performs late render phase actions, including scaling the renderer
// for the UI, converting UI draw commands to vertex buffer draw commands,
// passing the vertex buffers to the renderer, and presenting the renderer.
// PostRender should be called once at the end of every frame.
func (d *Driver) PostRender() (err error) {
	scaleX, scaleY := d.renderer.GetScale()
	if err = d.renderer.SetScale(d.uiScale, d.uiScale); err != nil {
		return fmt.Errorf("setting renderer scale to %g: %w", d.uiScale, err)
	}
	defer func() {
		if err2 := d.renderer.SetScale(scaleX, scaleY); err2 != nil && err == nil {
			err = fmt.Errorf("restoring renderer scale to %g x %g: %w", scaleX, scaleY, err)
		}
	}()
	d.commands.Clear()
	d.elements.Clear()
	d.vertices.Clear()
	if err = d.context.Convert(d.commands, d.vertices, d.elements, d.convertConf); err != nil {
		return fmt.Errorf("converting render commands: %w", err)
	}
	oldClipRect := d.renderer.GetClipRect()
	viewport := d.renderer.GetViewport()
	indices := reinterpretSlice[int32](d.elements.Memory(), 4)
	vertices := reinterpretSlice[sdl.Vertex](d.vertices.Memory(), int(vertexSize))
	d.context.DrawForEach(d.commands, func(cmd *nk.DrawCommand) bool {
		if cmd.ElemCount == 0 {
			return true
		}

		clipRect := sdl.Rect{
			X: int32(cmd.ClipRect.X),
			Y: int32(cmd.ClipRect.Y),
			W: int32(cmd.ClipRect.W),
			H: int32(cmd.ClipRect.H),
		}

		if d.clampClipRect {
			if clipRect.X < 0 {
				clipRect.W += clipRect.H
				clipRect.X = 0
			}
			if clipRect.Y < 0 {
				clipRect.H += clipRect.Y
				clipRect.Y = 0
			}
			if clipRect.W > viewport.W {
				clipRect.W = viewport.W
			}
			if clipRect.H > viewport.H {
				clipRect.H = viewport.H
			}
		}

		if err = d.renderer.SetClipRect(&clipRect); err != nil {
			err = fmt.Errorf("setting renderer clip rectangle: %w", err)
			return false
		}
		texture := handleToTexture(cmd.Texture)
		if err = d.renderer.RenderGeometry(texture, vertices, indices[:cmd.ElemCount]); err != nil {
			err = fmt.Errorf("rendering raw geometry: %w", err)
			return false
		}
		indices = indices[cmd.ElemCount:]
		return true
	})
	if err != nil {
		return fmt.Errorf("error in context.DrawForEach: %w", err)
	}
	if err := d.renderer.SetClipRect(&oldClipRect); err != nil {
		return fmt.Errorf("restoring clip rect: %w", err)
	}
	d.renderer.Present()
	return nil
}

// Destroy fress resources used by the Driver. Destroy should be called once
// in the lifetime of a Driver, after the last call to PostRender.
func (d *Driver) Destroy() (err error) {
	defer sdl.Quit()
	defer func() {
		if d.window != nil {
			if err2 := d.window.Destroy(); err2 != nil && err == nil {
				err = err2
			}
		}
	}()
	defer func() {
		if d.renderer != nil {
			if err2 := d.renderer.Destroy(); err2 != nil && err == nil {
				err = err2
			}
		}
	}()
	defer func() {
		if d.fontTex != nil {
			if err2 := d.fontTex.Destroy(); err2 != nil && err == nil {
				err = err2
			}
		}
	}()
	// all of the following calls are nil-safe
	defer d.context.Free()
	defer d.atlas.Free()
	defer d.convertConf.Free()
	defer d.commands.Free()
	defer d.elements.Free()
	defer d.vertices.Free()
	return nil
}

func (d *Driver) bakeFont() (nk.DrawNullTexture, error) {
	image, width, height := d.atlas.Bake(nk.FontAtlasRGBA32)
	if image == nil {
		return nk.DrawNullTexture{}, errors.New("font baking returned nil image")
	}
	var err error
	d.fontTex, err = d.renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STATIC, width, height)
	if err != nil {
		return nk.DrawNullTexture{}, fmt.Errorf("creating font texture: %w", err)
	}
	if err = d.fontTex.Update(nil, image, int(4*width)); err != nil {
		return nk.DrawNullTexture{}, fmt.Errorf("uploading font atlas to texture: %w", err)
	}
	if err = d.fontTex.SetBlendMode(sdl.BLENDMODE_BLEND); err != nil {
		return nk.DrawNullTexture{}, fmt.Errorf("setting texture blend mode: %w", err)
	}
	d.font.ScaleHeight(d.uiScale)
	null := d.atlas.End(textureToHandle(d.fontTex))
	d.atlas.Cleanup()
	return null, nil
}

func (d *Driver) computeUIScale() error {
	renderW, renderH, err := d.renderer.GetOutputSize()
	if err != nil {
		return fmt.Errorf("getting renderer output size: %w", err)
	}
	windowW, windowH := d.window.GetSize()
	renderScaleX := float32(renderW) / float32(windowW)
	renderScaleY := float32(renderH) / float32(windowH)
	if renderScaleY != renderScaleX {
		sdl.LogWarn(sdl.LOG_CATEGORY_APPLICATION,
			"display is scaled inconsistently (%f x %f)",
			renderScaleX, renderScaleY)
	}
	d.uiScale = renderScaleY
	return nil
}

type errQuit struct{}

func (errQuit) Error() string {
	return "application quit requested by the user"
}

// ErrQuit is an error sentinel value used to indicate that the application
// should quit.
var ErrQuit = errQuit{}

// EventListener is the function signature for the optional event listener,
// which is called after Nuklear handles an event. See the EventHandler type
// for a description of the other parameters.
type EventListener func(event sdl.Event, eventType EventType, usedByNuklear bool) error

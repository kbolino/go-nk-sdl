package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/kbolino/go-nk"
	"github.com/kbolino/go-nksdl"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	flagFont  = flag.String("font", "", "load font from file path (otherwise use built-in font)")
	flagHiDPI = flag.Bool("hiDPI", false, "enable high-DPI display support")
	flagVsync = flag.Bool("vsync", false, "enable sync on vertical blank (VSYNC)")
)

func init() {
	runtime.LockOSThread()
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}

func run() (err error) {
	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_DEBUG)
	windowFlags := uint32(0)
	if *flagHiDPI {
		sdl.SetHint(sdl.HINT_VIDEO_HIGHDPI_DISABLED, "0")
		windowFlags |= sdl.WINDOW_ALLOW_HIGHDPI
	}
	renderFlags := uint32(0)
	if *flagVsync {
		renderFlags = sdl.RENDERER_PRESENTVSYNC
	}
	sdlDriver := nksdl.DefaultSDLDriver{
		InitFlags: sdl.INIT_EVERYTHING,
		Window: nksdl.WindowOpts{
			Title:  "go-nk-sdl demo",
			PosX:   sdl.WINDOWPOS_CENTERED,
			PosY:   sdl.WINDOWPOS_CENTERED,
			Width:  800,
			Height: 600,
			Flags:  windowFlags,
		},
		Render: nksdl.RenderOpts{
			Flags: renderFlags,
		},
	}
	nkDriver := nksdl.DefaultNkDriver{
		Font: nksdl.FontOpts{
			Size: 13,
		},
		Convert: nksdl.ConvertOpts{
			GlobalAlpha:        1,
			LineAA:             nk.AntiAliasingOn,
			ShapeAA:            nk.AntiAliasingOn,
			CircleSegmentCount: nk.DefaultSegmentCount,
			CurveSegmentCount:  nk.DefaultSegmentCount,
			ArcSegmentCount:    nk.DefaultSegmentCount,
		},
	}
	driver := nksdl.NewDriver(&sdlDriver, &nkDriver, nksdl.DefaultBindings, nil)
	if err := driver.Init(); err != nil {
		return fmt.Errorf("initializing NkSDL driver: %w", err)
	}
	defer func() {
		if err := driver.Destroy(); err != nil {
			sdl.LogWarn(sdl.LOG_CATEGORY_APPLICATION, "error destroying NkSDL driver: %s", err.Error())
		}
	}()
	nkc := driver.Context()
	if err := driver.SetUIScale(0); err != nil {
		return fmt.Errorf("setting UI scale: %w", err)
	}
	driver.SetBGColor(sdl.Color{R: 31, B: 31, G: 31, A: 255})
	checked := false
	for {
		if err := driver.FrameStart(); err == nksdl.ErrQuit {
			break
		} else if err != nil {
			return fmt.Errorf("error in driver.FrameStart: %w", err)
		}

		// conventional rendering operations would go here

		if err := driver.PreGUI(); err != nil {
			return fmt.Errorf("error in driver.PreGUI: %w", err)
		}

		if nkc.Begin("Demo",
			&nk.Rect{X: 50, Y: 50, W: 230, H: 250},
			nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowMinimizable|nk.WindowTitle,
		) {
			nkc.LayoutRowStatic(30, 81, 1)
			if nkc.ButtonText("Button") {
				fmt.Println("button pressed")
			}
			nkc.LayoutRowStatic(30, 80, 1)
			checked = nkc.CheckText("Check me", checked)
		}
		nkc.End()

		if err := driver.FrameEnd(); err != nil {
			return fmt.Errorf("error in driver.FrameEnd: %w", err)
		}
	}
	return nil
}

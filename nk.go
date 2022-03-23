package nksdl

import (
	"unsafe"

	"github.com/kbolino/go-nk"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	vertexSize      = unsafe.Sizeof(sdl.Vertex{})
	vertexAlignment = unsafe.Alignof(sdl.Vertex{})
	vertexLayout    = []nk.DrawVertexLayoutElement{
		{Attribute: nk.VertexPosition, Format: nk.FormatFloat, Offset: unsafe.Offsetof(sdl.Vertex{}.Position)},
		{Attribute: nk.VertexColor, Format: nk.FormatR8G8B8A8, Offset: unsafe.Offsetof(sdl.Vertex{}.Color)},
		{Attribute: nk.VertexTexcoord, Format: nk.FormatFloat, Offset: unsafe.Offsetof(sdl.Vertex{}.TexCoord)},
	}
)

// NkDriver is implemented by any type capable of initializing Nuklear and its
// core resources.
type NkDriver interface {
	CreateContext() (*nk.Context, error)
	CreateFontAtlas() (*nk.FontAtlas, error)
	CreateFont(atlas *nk.FontAtlas, scale float32) (*nk.Font, error)
	CreateConvertConfig(
		vertexLayout []nk.DrawVertexLayoutElement,
		vertexSize, vertexAlignment uint32,
		null nk.DrawNullTexture,
	) *nk.ConvertConfig
}

// DefaultNkDriver is the default implementation of NkDriver. It can be used
// directly as an NkDriver or extended to patch/override its behavior.
type DefaultNkDriver struct {
	// Font contains options for creating fonts.
	Font FontOpts
	// Convert contains options for creating vertex buffer conversion
	// configurations.
	Convert ConvertOpts
}

var _ NkDriver = &DefaultNkDriver{}

func (d *DefaultNkDriver) CreateContext() (*nk.Context, error) {
	return nk.NewContext()
}

func (d *DefaultNkDriver) CreateFontAtlas() (*nk.FontAtlas, error) {
	return nk.NewFontAtlas(), nil
}

func (d *DefaultNkDriver) CreateFont(atlas *nk.FontAtlas, scale float32) (*nk.Font, error) {
	if d.Font.Path == "" {
		return atlas.AddDefaultFont(d.Font.Size*scale, nil), nil
	} else {
		return atlas.AddFromFile(d.Font.Path, d.Font.Size*scale, nil)
	}
}

func (d *DefaultNkDriver) CreateConvertConfig(
	vertexLayout []nk.DrawVertexLayoutElement,
	vertexSize, vertexAlignment uint32,
	null nk.DrawNullTexture,
) *nk.ConvertConfig {
	return nk.ConvertConfigBuilder{
		GlobalAlpha:        d.Convert.GlobalAlpha,
		LineAA:             d.Convert.LineAA,
		ShapeAA:            d.Convert.ShapeAA,
		CircleSegmentCount: d.Convert.CircleSegmentCount,
		CurveSegmentCount:  d.Convert.CurveSegmentCount,
		ArcSegmentCount:    d.Convert.ArcSegmentCount,
		VertexLayout:       vertexLayout,
		VertexSize:         vertexSize,
		VertexAlignment:    vertexAlignment,
		Null:               null,
	}.Build()
}

// ConvertOpts contains options used by DefaultNkContext.CreateConvertConfig.
type ConvertOpts struct {
	GlobalAlpha        float32
	LineAA, ShapeAA    nk.AntiAliasing
	CircleSegmentCount uint32
	CurveSegmentCount  uint32
	ArcSegmentCount    uint32
}

// FontOpts contains options used by DefaultNkContext.CreateFont.
type FontOpts struct {
	Path string
	Size float32
}

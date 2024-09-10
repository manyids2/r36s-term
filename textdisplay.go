package main

import (
	"image/color"
	"log"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"

	_ "embed"
)

//go:embed Lato-Bold.ttf
var latoBold []byte

type TextDisplay struct {
	x, y    int32
	color   color.NRGBA
	font    *ttf.Font
	text    string
	texture *sdl.Texture
	surface *sdl.Surface
}

func (t *TextDisplay) SetText(text string) error {
	t.Close()
	t.text = text
	surface, err := t.font.RenderUTF8Blended(t.text, sdl.Color(t.color))
	if err != nil {
		log.Printf("Cannot set text: %s", err)
		return err
	}
	t.surface = surface
	return nil
}

func (t *TextDisplay) Render(renderer *sdl.Renderer) {
	if t.surface == nil {
		return
	}
	if t.texture == nil {
		texture, err := renderer.CreateTextureFromSurface(t.surface)
		if err != nil {
			log.Printf("cannot render text: %s", err)
			return
		}
		t.texture = texture
	}
	r := sdl.Rect{X: t.x, Y: t.y, W: t.surface.W, H: t.surface.H}
	renderer.Copy(t.texture, nil, &r)
}

func (t *TextDisplay) Close() error {
	if t.texture != nil {
		t.texture.Destroy()
		t.texture = nil
	}
	if t.surface != nil {
		t.surface.Free()
		t.surface = nil
	}
	return nil
}

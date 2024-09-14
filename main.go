package main

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"os/exec"
	"strings"
	"time"

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

func runCmd(c string) ([]string, []string) {
	var cmd *exec.Cmd
	args := strings.Split(c, " ")
	if len(args) > 1 {
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		cmd = exec.Command(args[0])
	}
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	cmd.Run()
	out := outb.String()
	err := errb.String()
	return strings.Split(out, "\n"), strings.Split(err, "\n")
}

func main() {
	var font *ttf.Font

	if err := ttf.Init(); err != nil {
		return
	}
	defer ttf.Quit()

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("term", 0, 0, 640, 480, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Printf("Couldn't get accelerated renderer: %s", err)
		renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_SOFTWARE)
		if err != nil {
			panic(err)
		}
	}
	defer renderer.Destroy()

	sdl.ShowCursor(sdl.DISABLE)

	sdl.JoystickEventState(sdl.ENABLE)

	if rinfo, err := renderer.GetInfo(); err == nil {
		log.Printf("Renderer info: %#v", rinfo)
	}

	fontOps, err := sdl.RWFromMem(latoBold)
	if err != nil {
		panic(err)
	}
	// Load the font for our text
	if font, err = ttf.OpenFontRW(fontOps, 0, 12); err != nil {
		panic(err)
	}
	defer font.Close()
	defer fontOps.Close()

	// --- GUI ---
	cmd_text := TextDisplay{x: 10, y: 30, color: color.NRGBA{200, 200, 200, 255}, font: font}
	cmd_text.SetText(">")

	err_text := TextDisplay{x: 10, y: 460, color: color.NRGBA{200, 0, 0, 255}, font: font}
	err_text.SetText(".")

	out_texts := []TextDisplay{}
	for i := 0; i < 40; i++ {
		y := int32(50 + (i * 15))
		if y >= 440 {
			break
		}
		out_texts = append(out_texts, TextDisplay{x: 10, y: y, color: color.NRGBA{200, 200, 200, 255}, font: font})
		out_texts[i].SetText("---")
	}

	running := true
	tick := time.Tick(time.Microsecond * 33333)

	var joysticks [16]*sdl.Joystick
	text := ""

	sdl.StartTextInput()
	for running {
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		// Render everything
		cmd_text.Render(renderer)
		err_text.Render(renderer)
		for _, o := range out_texts {
			o.Render(renderer)
		}

		renderer.Present()

		<-tick

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {

			switch t := event.(type) {
			case *sdl.QuitEvent: // NOTE: Please use `*sdl.QuitEvent` for `v0.4.x` (current version).
				println("Quit")
				running = false
			case *sdl.KeyboardEvent:
				if t.State == sdl.PRESSED {
					if t.Keysym.Sym == sdl.K_ESCAPE {
						println("Quit")
						running = false
					}
					if t.Keysym.Sym == sdl.K_BACKSPACE {
						if len(text) > 0 {
							text = text[:len(text)-1]
						}
						if len(text) > 0 {
							cmd_text.SetText(text)
						} else {
							cmd_text.SetText(">")
						}
						println("in: ", text)
					}
					if t.Keysym.Sym == sdl.K_RETURN {
						if (t.Keysym.Mod == sdl.KMOD_LSHIFT) || (t.Keysym.Mod == sdl.KMOD_RSHIFT) {
							text = ""
							sdl.StartTextInput()
							println("Starting text capture")
							cmd_text.SetText("Enter command...")
						} else {
							sdl.StopTextInput()
							if len(text) > 0 {
								cmd_text.SetText(text)
							} else {
								cmd_text.SetText(">")
							}
							out, err := runCmd(text)
							println("text: ", text)
							for i, _ := range out_texts {
								out_texts[i].SetText("---")
							}
							for i, o := range out {
								if i < len(out_texts) {
									out_texts[i].SetText(o)
								}
							}
							fmt.Println("out: ", out)
							fmt.Println("err: ", err)
							if len(err) > 0 {
								err_text.SetText(err[0])
							}
						}
					}
				}
			case *sdl.TextInputEvent:
				text += event.(*sdl.TextInputEvent).GetText()
				if len(text) > 0 {
					cmd_text.SetText(text)
				} else {
					cmd_text.SetText(">")
				}
				println("in: ", text)
			case *sdl.JoyButtonEvent:
				if t.State == sdl.PRESSED {
				}
			case *sdl.JoyDeviceAddedEvent:
				// Open joystick for use
				joysticks[int(t.Which)] = sdl.JoystickOpen(int(t.Which))
				if joysticks[int(t.Which)] != nil {
					fmt.Println("Joystick", t.Which, "connected")
				}
			case *sdl.JoyDeviceRemovedEvent:
				if joystick := joysticks[int(t.Which)]; joystick != nil {
					joystick.Close()
				}
				fmt.Println("Joystick", t.Which, "disconnected")
			}
		}
	}

	sdl.StopTextInput()
}

package chip8

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"image/color"
	"time"
)

// video embeds a pixelgl window.
type video struct {
	gfx      [64 * 32]byte
	window   *pixelgl.Window
	keyMap   map[uint16]pixelgl.Button
	keysDown [16]*time.Ticker
}

func newVideo() (*video, error) {
	config := pixelgl.WindowConfig{
		Title:  "chip-8 emulator",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}

	w, err := pixelgl.NewWindow(config)
	if err != nil {
		return nil, err
	}

	mapping := map[uint16]pixelgl.Button{
		0x1: pixelgl.Key1, 0x2: pixelgl.Key2,
		0x3: pixelgl.Key3, 0xC: pixelgl.Key4,
		0x4: pixelgl.KeyQ, 0x5: pixelgl.KeyW,
		0x6: pixelgl.KeyE, 0xD: pixelgl.KeyR,
		0x7: pixelgl.KeyA, 0x8: pixelgl.KeyS,
		0x9: pixelgl.KeyD, 0xE: pixelgl.KeyF,
		0xA: pixelgl.KeyZ, 0x0: pixelgl.KeyX,
		0xB: pixelgl.KeyC, 0xF: pixelgl.KeyV,
	}

	v := &video{
		gfx:      [64 * 32]byte{},
		window:   w,
		keyMap:   mapping,
		keysDown: [16]*time.Ticker{},
	}

	return v, nil
}

func (v *video) draw() {
	v.window.Clear(color.Black)
	img := imdraw.New(nil)
	img.Color = pixel.RGB(1, 1, 1)

	var w float64 = 1024 / 64
	var h float64 = 768 / 32

	for i := 0; i < 64; i++ {
		for j := 0; j < 32; j++ {
			if v.gfx[(31-j)*64+i] == 0 {
				continue
			}

			img.Push(pixel.V(w*float64(i), h*float64(j)))
			img.Push(pixel.V(w*float64(i)+w, h*float64(j)+h))
			img.Rectangle(0)
		}
	}

	img.Draw(v.window)
	v.window.Update()
}

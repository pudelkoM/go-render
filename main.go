package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/pudelkoM/go-render/pkg/blockworld"
)

func init() {
	// GLFW: This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

func drawRandomPix(img *image.RGBA, _ *blockworld.Blockworld, _ int64) {
	for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
		img.Set(x, x, color.White)
	}

	rx := rand.Intn(img.Rect.Max.X)
	ry := rand.Intn(img.Rect.Max.Y)
	c := color.NRGBA{
		uint8(rand.Intn(256)),
		uint8(rand.Intn(256)),
		uint8(rand.Intn(256)),
		255,
	}
	img.Set(rx, ry, c)
}

func handleInputs(w *glfw.Window, world *blockworld.Blockworld) {
	if w.GetKey(glfw.KeyEscape) == glfw.Press {
		w.SetShouldClose(true)
	}
	const speed = 0.1
	if w.GetKey(glfw.KeyA) == glfw.Press || w.GetKey(glfw.KeyA) == glfw.Repeat {
		world.PlayerPos.X -= speed
	}
	if w.GetKey(glfw.KeyS) == glfw.Press || w.GetKey(glfw.KeyS) == glfw.Repeat {
		world.PlayerPos.Y -= speed
	}
	if w.GetKey(glfw.KeyD) == glfw.Press || w.GetKey(glfw.KeyD) == glfw.Repeat {
		world.PlayerPos.X += speed
	}
	if w.GetKey(glfw.KeyW) == glfw.Press || w.GetKey(glfw.KeyW) == glfw.Repeat {
		world.PlayerPos.Y += speed
	}
}

func renderBuf(img *image.RGBA, world *blockworld.Blockworld, frameCount int64) {
	// clear image
	draw.Draw(img, img.Rect, image.NewUniform(color.Black), image.ZP, draw.Src)

	pos := world.PlayerPos
	imgRatio := float64(img.Rect.Dy()) / float64(img.Rect.Dx())
	// fovHBlocks := 30.
	fovHBlocks := float64(img.Rect.Dx()) / float64(world.BlockSizePx)
	fovVBlocks := float64(fovHBlocks) / imgRatio
	// fmt.Println("fovHBlocks", fovHBlocks, "fovVBlocks", fovVBlocks)

	_, fracX := math.Modf(pos.X)
	_, fracY := math.Modf(pos.Y)
	subX := int(math.Round(fracX * float64(world.BlockSizePx)))
	subY := int(math.Round(fracY * float64(world.BlockSizePx)))

	// if frameCount%60 == 0 {
	// 	fmt.Println("fracX", fracX, "fracY", fracY, "subX", subX, "subY", subY)
	// }

	for x := 0; x < int(fovVBlocks); x++ {
		for y := 0; y < int(fovHBlocks); y++ {
			z := 0
			p := blockworld.Point{X: x, Y: y, Z: z}
			t := blockworld.PointFromVec(pos).Add(p)
			b, ok := world.Get(t)
			if !ok {
				continue
			}

			r := image.Rect(
				x*world.BlockSizePx-subX,
				y*world.BlockSizePx-subY,
				(x+1)*world.BlockSizePx-subX,
				(y+1)*world.BlockSizePx-subY,
			)
			draw.Draw(img, r, image.NewUniform(b.Color), image.ZP, draw.Src)
		}
	}
}

func main() {
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.DoubleBuffer, glfw.True)
	glfw.WindowHint(glfw.FocusOnShow, glfw.True)
	window, err := glfw.CreateWindow(640, 480, "My Window", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	glfw.SwapInterval(1)

	err = gl.Init()
	if err != nil {
		panic(err)
	}

	// window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	// 	if key == glfw.KeyEscape && action == glfw.Press {
	// 		w.SetShouldClose(true)
	// 	}
	// 	if key == glfw.KeyA && (action == glfw.Press || action == glfw.Repeat) {
	// 		fmt.Println("A")
	// 	}
	// 	if key == glfw.KeyS && (action == glfw.Press || action == glfw.Repeat) {
	// 		fmt.Println("S")
	// 	}
	// 	if key == glfw.KeyD && (action == glfw.Press || action == glfw.Repeat) {
	// 		fmt.Println("D")
	// 	}
	// 	if key == glfw.KeyW && (action == glfw.Press || action == glfw.Repeat) {
	// 		fmt.Println("W")
	// 	}
	// })

	var texture uint32
	{
		gl.GenTextures(1, &texture)

		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)

		// gl.BindImageTexture(0, texture, 0, false, 0, gl.WRITE_ONLY, gl.RGBA8)
	}

	var framebuffer uint32
	{
		gl.GenFramebuffers(1, &framebuffer)
		gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0)

		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, framebuffer)
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
	}

	// World setup
	var w, h = window.GetFramebufferSize()

	var img = image.NewRGBA(image.Rect(0, 0, w, h))
	fmt.Println("frame size", img.Rect)

	world := blockworld.NewBlockworld()
	world.Randomize()

	var frameCount int64 = 0
	var lastFrame = time.Now()

	for !window.ShouldClose() {
		handleInputs(window, world)
		renderBuf(img, world, frameCount)

		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(w), int32(h), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))

		gl.BlitFramebuffer(0, 0, int32(w), int32(h), 0, 0, int32(w), int32(h), gl.COLOR_BUFFER_BIT, gl.LINEAR)

		window.SwapBuffers()
		glfw.PollEvents()

		frameCount++
		took := time.Since(lastFrame)
		if frameCount%60 == 0 {
			fmt.Println("Frametime", took)
		}
		lastFrame = time.Now()
	}
}

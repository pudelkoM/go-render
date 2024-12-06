package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"runtime"
	"time"

	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/pudelkoM/go-render/pkg/blockworld"
	"github.com/pudelkoM/go-render/pkg/maploader"
)

func init() {
	// GLFW: This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

func handleInputs(w *glfw.Window, world *blockworld.Blockworld) {
	if w.GetKey(glfw.KeyEscape) == glfw.Press {
		w.SetShouldClose(true)
	}
	const speed = 0.3
	const rotSpeed = 3.
	if w.GetKey(glfw.KeyA) == glfw.Press || w.GetKey(glfw.KeyA) == glfw.Repeat {
		world.PlayerPos = world.PlayerPos.Add(world.PlayerDir.RotatePhi(-90).ToCartesianVec3(speed))
	}
	if w.GetKey(glfw.KeyS) == glfw.Press || w.GetKey(glfw.KeyS) == glfw.Repeat {
		world.PlayerPos = world.PlayerPos.Sub(world.PlayerDir.ToCartesianVec3(speed))
	}
	if w.GetKey(glfw.KeyD) == glfw.Press || w.GetKey(glfw.KeyD) == glfw.Repeat {
		world.PlayerPos = world.PlayerPos.Add(world.PlayerDir.RotatePhi(90).ToCartesianVec3(speed))
	}
	if w.GetKey(glfw.KeyW) == glfw.Press || w.GetKey(glfw.KeyW) == glfw.Repeat {
		world.PlayerPos = world.PlayerPos.Add(world.PlayerDir.ToCartesianVec3(speed))
	}
	if w.GetKey(glfw.KeyUp) == glfw.Press || w.GetKey(glfw.KeyUp) == glfw.Repeat {
		world.PlayerDir.Theta += rotSpeed
		world.PlayerDir = world.PlayerDir.Normalize()
	}
	if w.GetKey(glfw.KeyDown) == glfw.Press || w.GetKey(glfw.KeyDown) == glfw.Repeat {
		world.PlayerDir.Theta -= rotSpeed
		world.PlayerDir = world.PlayerDir.Normalize()
	}
	if w.GetKey(glfw.KeyLeft) == glfw.Press || w.GetKey(glfw.KeyLeft) == glfw.Repeat {
		world.PlayerDir = world.PlayerDir.RotatePhi(-rotSpeed)
	}
	if w.GetKey(glfw.KeyRight) == glfw.Press || w.GetKey(glfw.KeyRight) == glfw.Repeat {
		world.PlayerDir = world.PlayerDir.RotatePhi(rotSpeed)
	}
}

func renderBuf(img *image.RGBA, world *blockworld.Blockworld, frameCount int64) {
	// clear image
	draw.Draw(img, img.Rect, image.NewUniform(color.Black), image.ZP, draw.Src)

	imgRatio := float64(img.Rect.Dy()) / float64(img.Rect.Dx())
	fovHDeg := 55.
	fovVDeg := fovHDeg * imgRatio
	degPerPixel := fovHDeg / float64(img.Rect.Dx())

	for x := 0; x < img.Rect.Dx(); x++ {
		xd := (-fovHDeg / 2) + float64(x)*degPerPixel
		for y := 0; y < img.Rect.Dy(); y++ {
			yd := (-fovVDeg / 2) + float64(y)*degPerPixel
			rayVec := blockworld.Angle3{Theta: world.PlayerDir.Theta + yd, Phi: world.PlayerDir.Phi + xd}.ToCartesianVec3(1)
			newPos := world.PlayerPos
			for i := 0; i < 150; i++ {
				newPos = newPos.Add(rayVec)
				n := newPos.ToPointTrunc()
				b, ok := world.Get(n)
				if !ok {
					continue
				}
				// fmt.Println("found block at ", n, " color ", b.Color)
				img.Set(x, img.Rect.Dy()-y, b.Color) // flip y coord because ogl texture use bottom-left as origin
				break
			}
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

	const renderScale = 16
	var w, h = window.GetFramebufferSize()
	w /= renderScale
	h /= renderScale
	var img = image.NewRGBA(image.Rect(0, 0, w, h))
	fmt.Println("frame size", img.Rect)

	// World setup
	world := blockworld.NewBlockworld()
	// err = maploader.LoadMap("./maps/AttackonDeuces.vxl", world)
	err = maploader.LoadMap("./maps/DragonsReach.vxl", world)
	if err != nil {
		panic(err)
	}

	var frameCount int64 = 0
	var lastFrame = time.Now()

	for !window.ShouldClose() {
		handleInputs(window, world)
		renderBuf(img, world, frameCount)

		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(w), int32(h), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))

		gl.BlitFramebuffer(0, 0, int32(w), int32(h), 0, 0, int32(w)*renderScale, int32(h)*renderScale, gl.COLOR_BUFFER_BIT, gl.NEAREST)

		window.SwapBuffers()
		glfw.PollEvents()

		frameCount++
		took := time.Since(lastFrame)
		if frameCount%60 == 0 {
			fmt.Println("Frametime", took, "FPS", 1/took.Seconds())
			fmt.Println("PlayerPos", world.PlayerPos, "PlayerDir", world.PlayerDir)
		}
		lastFrame = time.Now()
	}
}

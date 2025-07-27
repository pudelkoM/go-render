package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/pudelkoM/go-render/pkg/blockworld"
	"github.com/pudelkoM/go-render/pkg/maploader"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func init() {
	// GLFW: This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

const (
	renderNormal = iota
	renderDepth
)

var (
	mapIndex   = 0
	renderMode = renderNormal
)

func handleInputs(w *glfw.Window, world *blockworld.Blockworld) {
	if w.GetKey(glfw.KeyEscape) == glfw.Press {
		w.SetShouldClose(true)
	}
	const speed = 0.3
	const rotSpeed = 3.
	if w.GetKey(glfw.KeyA) == glfw.Press || w.GetKey(glfw.KeyA) == glfw.Repeat {
		world.PlayerPos = world.PlayerPos.Add(world.PlayerDir.ResetTheta().RotatePhi(-90).ToCartesianVec3(speed))
	}
	if w.GetKey(glfw.KeyS) == glfw.Press || w.GetKey(glfw.KeyS) == glfw.Repeat {
		world.PlayerPos = world.PlayerPos.Sub(world.PlayerDir.ToCartesianVec3(speed))
	}
	if w.GetKey(glfw.KeyD) == glfw.Press || w.GetKey(glfw.KeyD) == glfw.Repeat {
		world.PlayerPos = world.PlayerPos.Add(world.PlayerDir.ResetTheta().RotatePhi(90).ToCartesianVec3(speed))
	}
	if w.GetKey(glfw.KeyW) == glfw.Press || w.GetKey(glfw.KeyW) == glfw.Repeat {
		world.PlayerPos = world.PlayerPos.Add(world.PlayerDir.ToCartesianVec3(speed))
	}
	if w.GetKey(glfw.KeyQ) == glfw.Press || w.GetKey(glfw.KeyQ) == glfw.Repeat {
		world.PlayerPos.Z += speed
	}
	if w.GetKey(glfw.KeyE) == glfw.Press || w.GetKey(glfw.KeyE) == glfw.Repeat {
		world.PlayerPos.Z -= speed
	}
	if w.GetKey(glfw.KeyUp) == glfw.Press || w.GetKey(glfw.KeyUp) == glfw.Repeat {
		world.PlayerDir.Theta += rotSpeed
		world.PlayerDir = world.PlayerDir.ClampToView()
	}
	if w.GetKey(glfw.KeyDown) == glfw.Press || w.GetKey(glfw.KeyDown) == glfw.Repeat {
		world.PlayerDir.Theta -= rotSpeed
		world.PlayerDir = world.PlayerDir.ClampToView()
	}
	if w.GetKey(glfw.KeyLeft) == glfw.Press || w.GetKey(glfw.KeyLeft) == glfw.Repeat {
		world.PlayerDir = world.PlayerDir.RotatePhi(-rotSpeed)
	}
	if w.GetKey(glfw.KeyRight) == glfw.Press || w.GetKey(glfw.KeyRight) == glfw.Repeat {
		world.PlayerDir = world.PlayerDir.RotatePhi(rotSpeed)
	}
	if w.GetKey(glfw.KeyN) == glfw.Press {
		dir := "./maps/"
		files, err := os.ReadDir(dir)
		if err != nil {
			log.Fatal(err)
		}
		files = slices.DeleteFunc(files, func(f os.DirEntry) bool {
			return !strings.HasSuffix(f.Name(), ".vxl")
		})
		mapIndex = ((mapIndex + 1) % len(files))
		fmt.Println("loading map", files[mapIndex].Name())
		err = maploader.LoadMap(dir+files[mapIndex].Name(), world)
		if err != nil {
			panic(err)
		}
	}
	if w.GetKey(glfw.KeyL) == glfw.Press {
		if renderMode == renderNormal {
			renderMode = renderDepth
		} else {
			renderMode = renderNormal
		}
	}
}

func castRayAmatidesWoo(img *image.RGBA, world *blockworld.Blockworld,
	x, y int, newPos, rayVec blockworld.Vec3) {
	stepX := 0
	tDeltaX := 0.0
	tMaxX := 0.0
	if rayVec.X > 0 {
		stepX = 1
		tDeltaX = 1 / rayVec.X
		tMaxX = (math.Floor(newPos.X+1) - newPos.X) / rayVec.X
	} else if rayVec.X < 0 {
		stepX = -1
		tDeltaX = 1 / -rayVec.X
		tMaxX = (math.Ceil(newPos.X-1) - newPos.X) / rayVec.X
	} else {
		tMaxX = math.Inf(1)
	}

	stepY := 0
	tDeltaY := 0.0
	tMaxY := 0.0
	if rayVec.Y > 0 {
		stepY = 1
		tDeltaY = 1 / rayVec.Y
		tMaxY = (math.Floor(newPos.Y+1) - newPos.Y) / rayVec.Y
	} else if rayVec.Y < 0 {
		stepY = -1
		tDeltaY = 1 / -rayVec.Y
		tMaxY = (math.Ceil(newPos.Y-1) - newPos.Y) / rayVec.Y
	} else {
		tMaxY = math.Inf(1)
	}

	stepZ := 0
	tDeltaZ := 0.0
	tMaxZ := 0.0
	if rayVec.Z > 0 {
		stepZ = 1
		tDeltaZ = 1 / rayVec.Z
		tMaxZ = (math.Floor(newPos.Z+1) - newPos.Z) / rayVec.Z
	} else if rayVec.Z < 0 {
		stepZ = -1
		tDeltaZ = 1 / -rayVec.Z
		tMaxZ = (math.Ceil(newPos.Z-1) - newPos.Z) / rayVec.Z
	} else {
		tMaxZ = math.Inf(1)
	}

	for i := 0; i < 250; i++ {
		if tMaxX < tMaxY && tMaxX < tMaxZ {
			// Idea: store signed distance to nearest block per block
			// in world map and use it to skip empty space faster.
			newPos.X += float64(stepX)
			tMaxX += tDeltaX
		} else if tMaxY < tMaxZ {
			newPos.Y += float64(stepY)
			tMaxY += tDeltaY
		} else {
			newPos.Z += float64(stepZ)
			tMaxZ += tDeltaZ
		}

		n := newPos.ToPointTrunc()
		b, ok := world.Get(n)
		if !ok {
			// Advance vector to next full block?
			continue
		}
		img.Set(x, y, b.Color)

		if renderMode == renderDepth {
			v := blockworld.MagmaClamp(float64(i) / 250.)
			c := color.RGBA{
				R: uint8(v.X * 255),
				G: uint8(v.Y * 255),
				B: uint8(v.Z * 255),
				A: 255,
			}
			img.Set(x, y, c)
		}
		break
	}
}

func renderBuf(img *image.RGBA, world *blockworld.Blockworld, frameCount int64,
	lastFrameDuration time.Duration) {
	// clear image
	draw.Draw(img, img.Rect, image.NewUniform(color.Black), image.Point{}, draw.Src)

	imgRatio := float64(img.Rect.Dy()) / float64(img.Rect.Dx())
	fovHDeg := 55.
	fovVDeg := fovHDeg * imgRatio
	degPerPixel := fovHDeg / float64(img.Rect.Dx())

	const threads = 4
	yDD := int(math.Ceil(float64(img.Rect.Dy()) / threads))
	wg := sync.WaitGroup{}
	wg.Add(threads)
	for t := 0; t < threads; t++ {
		go func(t int) {
			defer wg.Done()
			yStart := t * yDD
			if yStart >= img.Rect.Dy() {
				return
			}
			yEnd := (t + 1) * yDD
			if yEnd >= img.Rect.Dy() {
				yEnd = img.Rect.Dy()
			}

			for y := yStart; y < yEnd; y++ {
				yd := (-fovVDeg / 2) + float64(y)*degPerPixel
				for x := 0; x < img.Rect.Dx(); x++ {
					xd := (-fovHDeg / 2) + float64(x)*degPerPixel
					viewVec := blockworld.Vec3{X: 1, Y: 0, Z: 0}
					rayVec := viewVec.RotateY(yd).RotateZ(xd)
					rayVec = rayVec.RotateY(world.PlayerDir.Theta - 90).RotateZ(world.PlayerDir.Phi)
					newPos := world.PlayerPos

					castRayAmatidesWoo(img, world, x, y, newPos, rayVec)
				}
			}
		}(t)
	}
	wg.Wait()

	img.SetRGBA(img.Rect.Dx()/2, img.Rect.Dy()/2, color.RGBA{R: 255, A: 255})

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{G: 255, A: 255}),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(2, 12),
	}
	d.DrawString(fmt.Sprintf("FPS: %03.0f ", 1/lastFrameDuration.Seconds()))
	d.DrawString(fmt.Sprintf("Frame: %v ", frameCount))
	d.Dot = fixed.P(2, 24)
	d.DrawString(fmt.Sprintf("Pos: %v Dir: %v ", world.PlayerPos, world.PlayerDir))
}

func main() {
	go func() {
		log.Fatal(http.ListenAndServe(":6060", nil))
	}()

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
	window.SetInputMode(glfw.StickyKeysMode, glfw.True)

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

	const renderScale = 8
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
	// world.PlayerPos = blockworld.Vec3{X: 154, Y: 256.5, Z: 40}
	// world.PlayerDir = blockworld.Angle3{Theta: 0, Phi: 0}

	// Side view.
	world.PlayerPos = blockworld.Vec3{X: 190, Y: 310, Z: 33}
	world.PlayerDir = blockworld.Angle3{Theta: 95, Phi: 325}

	// Starting window.
	// world.PlayerPos = blockworld.Vec3{X: 154, Y: 256.5, Z: 40}
	// world.PlayerDir = blockworld.Angle3{Theta: 90, Phi: 0}

	var frameCount int64 = 0
	var lastFrame = time.Now()
	var lastFrameDuration time.Duration = 0
	// For stats printing only.
	var lastFrameCount int64 = 0
	var lastFrameTime = time.Now()

	for !window.ShouldClose() {
		handleInputs(window, world)
		renderBuf(img, world, frameCount, lastFrameDuration)

		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(w), int32(h), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))

		gl.BlitFramebuffer(
			0, int32(h), int32(w), 0,
			0, 0, int32(w)*renderScale, int32(h)*renderScale,
			gl.COLOR_BUFFER_BIT, gl.NEAREST)
		window.SwapBuffers()
		glfw.PollEvents()

		if lastFrame.Sub(lastFrameTime) > time.Second {
			deltaFrames := frameCount - lastFrameCount
			deltaTime := time.Since(lastFrameTime)
			fps := float64(deltaFrames) / deltaTime.Seconds()
			avgFrameTime := deltaTime / time.Duration(deltaFrames)
			fmt.Println("Frametime", avgFrameTime, "FPS", fps)
			fmt.Println("PlayerPos", world.PlayerPos, "PlayerDir", world.PlayerDir)
			lastFrameCount = frameCount
			lastFrameTime = lastFrame
		}
		frameCount++
		lastFrameDuration = time.Since(lastFrame)
		lastFrame = time.Now()
	}
}

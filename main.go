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

var (
	mapIndex = 0
)

func handleInputs(w *glfw.Window, world *blockworld.Blockworld) {
	if w.GetKey(glfw.KeyEscape) == glfw.Press {
		w.SetShouldClose(true)
	}
	const speed = 0.3
	const rotSpeed = 2.
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
	if w.GetKey(glfw.KeyMinus) == glfw.Press {
		world.PlayerFovHDeg -= 1.5
		if world.PlayerFovHDeg < 1 {
			world.PlayerFovHDeg = 1
		}
	}
	if w.GetKey(glfw.KeyEqual) == glfw.Press {
		world.PlayerFovHDeg += 1.5
		if world.PlayerFovHDeg > 359 {
			world.PlayerFovHDeg = 359
		}
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
}

func amatidesWoo(newPos, rayVec blockworld.Vec3, world *blockworld.Blockworld) blockworld.Vec3 {

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

	for i := 0; i < 550; i++ {

		if tMaxX < tMaxY && tMaxX < tMaxZ {
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

		if b == nil {
			// End of map reached, stop looking.
			return newPos
		}
		if !ok {
			// Advance vector to next full block?
			continue
		}
	}
	return newPos
}

func renderBuf(img *image.RGBA, world *blockworld.Blockworld, frameCount int64, lastFrame time.Time) {
	// clear image
	draw.Draw(img, img.Rect, image.NewUniform(color.Black), image.ZP, draw.Src)

	// depth buffer
	// depth := image.NewGray16(image.Rect(0, 0, img.Rect.Dx(), img.Rect.Dy()))
	// draw.Draw(depth, depth.Rect, image.NewUniform(color.Gray{Y: 0}), image.ZP, draw.Src)

	imgRatio := float64(img.Rect.Dy()) / float64(img.Rect.Dx())
	fovHDeg := world.PlayerFovHDeg
	fovVDeg := fovHDeg * imgRatio
	degPerPixel := fovHDeg / float64(img.Rect.Dx())

	const threads = 8
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
			// fmt.Println("t", t, "yStart", yStart, "yEnd", yEnd)

			for y := yStart; y < yEnd; y++ {
				yd := (-fovVDeg / 2) + float64(y)*degPerPixel
				for x := 0; x < img.Rect.Dx(); x++ {
					xd := (-fovHDeg / 2) + float64(x)*degPerPixel
					viewVec := blockworld.Vec3{X: 1, Y: 0, Z: 0}
					rayVec := viewVec.RotateY(yd).RotateZ(xd)
					rayVec = rayVec.RotateY(world.PlayerDir.Theta - 90).RotateZ(world.PlayerDir.Phi)
					_, c := world.RayMarchSdf(world.PlayerPos, rayVec)
					img.Set(x, y, c)
				}
			}
		}(t)

		// for y := 0; y < img.Rect.Dy(); y++ {
		// 	yd := (-fovVDeg / 2) + float64(y)*degPerPixel
		// 	go func(t int) {
		// 		defer wg.Done()
		// 		xStart := t * (img.Rect.Dx() / threads)
		// 		if xStart >= img.Rect.Dx() {
		// 			return
		// 		}
		// 		xEnd := (t + 1) * (img.Rect.Dx() / threads)

		// 		for x := xStart; x < xEnd && x < img.Rect.Dx(); x++ {
		// 			xd := (-fovHDeg / 2) + float64(x)*degPerPixel
		// 			viewVec := blockworld.Vec3{X: 1, Y: 0, Z: 0}
		// 			rayVec := viewVec.RotateY(yd).RotateZ(xd)
		// 			rayVec = rayVec.RotateY(world.PlayerDir.Theta - 90).RotateZ(world.PlayerDir.Phi)
		// 			newPos := world.RayMarchSdf(world.PlayerPos, rayVec)
		// 			n := newPos.ToPointTrunc()
		// 			b, _ := world.Get(n)
		// 			if b == nil {
		// 				// End of map reached, stop looking.
		// 				continue
		// 			}
		// 			img.Set(x, img.Rect.Dy()-y, b.Color) // flip y coord because ogl texture use bottom-left as origin
		// 		}
		// 	}(t)
		// }

		// for x := 0; x < img.Rect.Dx(); x++ {
		// 	xd := (-fovHDeg / 2) + float64(x)*degPerPixel
		// 	viewVec := blockworld.Vec3{X: 1, Y: 0, Z: 0}
		// 	rayVec := viewVec.RotateY(yd).RotateZ(xd)
		// 	rayVec = rayVec.RotateY(world.PlayerDir.Theta - 90).RotateZ(world.PlayerDir.Phi)
		// 	// isReflectionRay := false

		// 	newPos := world.RayMarchSdf(world.PlayerPos, rayVec)

		// 	{
		// 		n := newPos.ToPointTrunc()
		// 		b, _ := world.Get(n)
		// 		if b == nil {
		// 			// End of map reached, stop looking.
		// 			continue
		// 		}

		// 		// face := blockworld.GetBlockFace(newPos, n)
		// 		// // orig := b.Color
		// 		// switch face {
		// 		// case blockworld.BLOCK_FACE_TOP:
		// 		// 	rayVec.Z = -rayVec.Z
		// 		// 	newPos.Z += 0.001
		// 		// case blockworld.BLOCK_FACE_BOTTOM:
		// 		// 	rayVec.Z = -rayVec.Z
		// 		// 	newPos.Z -= 0.001
		// 		// case blockworld.BLOCK_FACE_LEFT:
		// 		// 	rayVec.X = -rayVec.X
		// 		// 	newPos.X -= 0.001
		// 		// case blockworld.BLOCK_FACE_RIGHT:
		// 		// 	rayVec.X = -rayVec.X
		// 		// 	newPos.X += 0.001
		// 		// case blockworld.BLOCK_FACE_FRONT:
		// 		// 	rayVec.Y = -rayVec.Y
		// 		// 	newPos.Y += 0.001
		// 		// case blockworld.BLOCK_FACE_BACK:
		// 		// 	rayVec.Y = -rayVec.Y
		// 		// 	newPos.Y -= 0.001
		// 		// default:
		// 		// 	img.Set(x, img.Rect.Dy()-y, b.Color) // flip y coord because ogl texture use bottom-left as origin
		// 		// }
		// 		// Reflected ray
		// 		// newPos := world.RayMarchSdf(newPos, rayVec)
		// 		// n = newPos.ToPointTrunc()
		// 		// bRef, _ := world.Get(n)
		// 		// if bRef != nil {
		// 		// 	c := utils.CompositeNRGBA(orig, bRef.Color)
		// 		// 	img.Set(x, img.Rect.Dy()-y, c) // flip y coord because ogl texture use bottom-left as origin
		// 		// } else {
		// 		img.Set(x, img.Rect.Dy()-y, b.Color) // flip y coord because ogl texture use bottom-left as origin
		// 		// }

		// 		// if b.Reflective {
		// 		// Reflect ray by inverting Z component
		// 		// isReflectionRay = true
		// 		// rayVec = blockworld.Vec3{X: rayVec.X, Y: rayVec.Y, Z: -rayVec.Z}
		// 		// newPos = newPos.Add(rayVec)
		// 		// stepZ *= -1
		// 		// tDeltaZ *= -1
		// 		// tMaxZ = (math.Ceil(newPos.Z-1) - newPos.Z) / rayVec.Z
		// 		// 	img.Set(x, img.Rect.Dy()-y, b.Color)
		// 		// 	continue
		// 		// }
		// 		// if isReflectionRay {
		// 		// 	c1 := img.At(x, img.Rect.Dy()-y).(color.RGBA) // Color of the block we reflected off
		// 		// 	c2 := b.Color
		// 		// 	c1.A = 200
		// 		// 	c := utils.CompositeNRGBA(c1, c2)
		// 		// 	img.Set(x, img.Rect.Dy()-y, c)
		// 		// 	break
		// 		// }

		// 		// d := newPos.Sub(world.PlayerPos).Length()
		// 		// depth.SetGray16(x, img.Rect.Dy()-y, color.Gray16{Y: uint16(0xffff - d*300)})

		// 		continue
		// 	}
		// }
	}
	wg.Wait()

	img.SetRGBA(img.Rect.Dx()/2, img.Rect.Dy()/2, color.RGBA{R: 255, A: 255})

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{G: 255, A: 255}),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(2, 12),
	}
	dt := time.Since(lastFrame)
	d.DrawString(fmt.Sprintf("FPS: %03.0f ", 1/dt.Seconds()))
	d.DrawString(fmt.Sprintf("Frame: %v ", frameCount))
	d.Dot = fixed.P(2, 24)
	d.DrawString(fmt.Sprintf("Pos: %v Dir: %v ", world.PlayerPos, world.PlayerDir))
	d.DrawString(fmt.Sprintf("Fov: %0.2f ", world.PlayerFovHDeg))

	// draw.Draw(img, img.Rect, depth, image.ZP, draw.Over)
	// draw.DrawMask(img, img.Rect, image.NewUniform(color.NRGBA{R: 255, A: 64}), image.ZP, depth, image.ZP, draw.Over)

	// draw.DrawMask(img, img.Rect, image.NewUniform(color.NRGBA{R: 255, A: 255}), image.ZP, depth, image.ZP, draw.Over)
}

func main() {
	go func() {
		log.Fatal(http.ListenAndServe(":6060", nil))
	}()

	// audioCtx := audio.InitAudio()

	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.DoubleBuffer, glfw.True)
	glfw.WindowHint(glfw.FocusOnShow, glfw.True)
	window, err := glfw.CreateWindow(1200, 720, "FWMC brainrot sim", nil, nil)
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

	var textureOverlay uint32
	{
		gl.GenTextures(1, &textureOverlay)

		gl.BindTexture(gl.TEXTURE_2D, textureOverlay)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	}

	var framebufferOverlay uint32
	{
		gl.GenFramebuffers(1, &framebufferOverlay)
		gl.BindFramebuffer(gl.FRAMEBUFFER, framebufferOverlay)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, textureOverlay, 0)

		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, framebufferOverlay)
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
	}

	const renderScale = 4
	var w, h = window.GetFramebufferSize()
	imgOverlay := image.NewRGBA(image.Rect(0, 0, w, h))
	w /= renderScale
	h /= renderScale
	var img = image.NewRGBA(image.Rect(0, 0, w, h))
	fmt.Println("frame size", img.Rect)

	// World setup
	world := blockworld.NewBlockworld()
	// err = maploader.LoadMap("./maps/AttackonDeuces.vxl", world)
	err = maploader.LoadMap("./maps/DragonsReach.vxl", world)
	// err = maploader.LoadMap("./maps/shigaichi4.vxl", world)
	if err != nil {
		panic(err)
	}

	// fuwa, err := utils.LoadPNG("./assets/fuwa_64.png")
	// if err != nil {
	// 	panic(err)
	// }
	// moco, err := utils.LoadPNG("./assets/moco_64.png")
	// if err != nil {
	// 	panic(err)
	// }
	// world.SetBlockTex(fuwa, moco)

	// blockworld.GenerateLabyrinth(world, 19, 19)

	// blockworld.GenerateLabyrinth(world, 31, 31)

	// world.PlayerPos = blockworld.Vec3{X: 154, Y: 256.5, Z: 40}
	// world.PlayerDir = blockworld.Angle3{Theta: 0, Phi: 0}

	// Side view.
	world.PlayerPos = blockworld.Vec3{X: 190, Y: 310, Z: 33}
	world.PlayerDir = blockworld.Angle3{Theta: 95, Phi: 325}

	// Starting window.
	// world.PlayerPos = blockworld.Vec3{X: 154, Y: 256.5, Z: 40}
	// world.PlayerDir = blockworld.Angle3{Theta: 90, Phi: 0}

	for i := imgOverlay.Rect.Min.X; i < imgOverlay.Rect.Max.X; i++ {
		imgOverlay.Set(i, i, color.RGBA{R: 255, A: 255})
	}

	var frameCount int64 = 0
	var lastFrame = time.Now()

	// solverState := &labsolver.LabSolver{}

	for !window.ShouldClose() {
		handleInputs(window, world)
		// labsolver.Advance(world, solverState, frameCount)
		// audio.HandleAudio(audioCtx, world, frameCount)
		renderBuf(img, world, frameCount, lastFrame)

		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(w), int32(h), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))

		// preBlit := time.Now()
		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, framebuffer)
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
		// gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, framebufferOverlay)
		gl.BlitFramebuffer(
			0, int32(h), int32(w), 0,
			0, 0, int32(w)*renderScale, int32(h)*renderScale,
			gl.COLOR_BUFFER_BIT, gl.NEAREST)
		// postBlit := time.Since(preBlit)
		// fmt.Println("Blit took", postBlit)

		// // Read window framebuffer
		// // gl.BindFramebuffer(gl.READ_FRAMEBUFFER, 0)
		// gl.BindFramebuffer(gl.READ_FRAMEBUFFER, framebufferOverlay)
		// gl.ReadPixels(0, 0, int32(w)*renderScale, int32(h)*renderScale, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(imgOverlay.Pix))

		// postBlit := time.Since(preBlit)
		// fmt.Println("dt took", postBlit)

		// // Overlay
		// d := &font.Drawer{
		// 	Dst:  imgOverlay,
		// 	Src:  image.NewUniform(color.RGBA{G: 255, A: 255}),
		// 	Face: basicfont.Face7x13,
		// 	Dot:  fixed.P(10, 20),
		// }
		// dt := time.Since(lastFrame)
		// d.DrawString(fmt.Sprintf("FPS: %0.0f", 1/dt.Seconds()))
		// gl.BindTexture(gl.TEXTURE_2D, textureOverlay)
		// gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(imgOverlay.Rect.Dx()), int32(imgOverlay.Rect.Dy()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(imgOverlay.Pix))

		// gl.BindFramebuffer(gl.READ_FRAMEBUFFER, framebufferOverlay)
		// gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
		// gl.BlitFramebuffer(
		// 	0, int32(imgOverlay.Rect.Dy()), int32(imgOverlay.Rect.Dx()), 0,
		// 	0, 0, int32(imgOverlay.Rect.Dx()), int32(imgOverlay.Rect.Dy()),
		// 	gl.COLOR_BUFFER_BIT, gl.NEAREST)

		// postBlit := time.Since(preBlit)
		// fmt.Println("total blit took", postBlit)

		window.SwapBuffers()
		glfw.PollEvents()

		frameCount++
		took := time.Since(lastFrame)
		if frameCount%60 == 0 {
			fmt.Println("Frametime", took, "FPS", 1/took.Seconds())
			fmt.Println("PlayerPos", world.PlayerPos, "PlayerDir", world.PlayerDir)
			b, _ := world.Get(world.PlayerPos.ToPointTrunc())
			if b != nil {
				// fmt.Printf("Player pos block %+v\n", b)
			}
		}
		lastFrame = time.Now()
	}
}

package blockworld_test

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/pudelkoM/go-render/pkg/blockworld"
	"github.com/pudelkoM/go-render/pkg/maploader"
)

func TestVec3_RotateZ(t *testing.T) {
	tests := []struct {
		name     string
		v        blockworld.Vec3
		angle    float64
		expected blockworld.Vec3
	}{
		{
			name:     "rotate 90 degrees",
			v:        blockworld.Vec3{X: 1, Y: 0, Z: 0},
			angle:    90,
			expected: blockworld.Vec3{X: 0, Y: 1, Z: 0},
		},
		{
			name:     "rotate 180 degrees",
			v:        blockworld.Vec3{X: 1, Y: 0, Z: 0},
			angle:    180,
			expected: blockworld.Vec3{X: -1, Y: 0, Z: 0},
		},
		{
			name:     "rotate 270 degrees",
			v:        blockworld.Vec3{X: 1, Y: 0, Z: 0},
			angle:    270,
			expected: blockworld.Vec3{X: 0, Y: -1, Z: 0},
		},
		{
			name:     "rotate 360 degrees",
			v:        blockworld.Vec3{X: 1, Y: 0, Z: 0},
			angle:    360,
			expected: blockworld.Vec3{X: 1, Y: 0, Z: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.RotateZ(tt.angle)
			if !almostEqual(result, tt.expected) {
				t.Errorf("RotateZ() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestAngle3_ToCartesianVec3(t *testing.T) {
	tests := []struct {
		name     string
		angle    blockworld.Angle3
		expected blockworld.Vec3
	}{
		{
			name:     "zero angles",
			angle:    blockworld.Angle3{Theta: 0, Phi: 0},
			expected: blockworld.Vec3{X: 0, Y: 0, Z: 1},
		},
		{
			name:     "90 degrees polar",
			angle:    blockworld.Angle3{Theta: 90, Phi: 0},
			expected: blockworld.Vec3{X: 1, Y: 0, Z: 0},
		},
		{
			name:     "45 degrees polar",
			angle:    blockworld.Angle3{Theta: 45, Phi: 0},
			expected: blockworld.Vec3{X: math.Sqrt2 / 2, Y: 0, Z: math.Sqrt2 / 2},
		},
		{
			name:     "90 degrees azimuthal",
			angle:    blockworld.Angle3{Theta: 0, Phi: 90},
			expected: blockworld.Vec3{X: 0, Y: 0, Z: 1},
		},
		{
			name:     "90 degrees polar 45 degrees azimuthal",
			angle:    blockworld.Angle3{Theta: 90, Phi: 45},
			expected: blockworld.Vec3{X: math.Sqrt2 / 2, Y: math.Sqrt2 / 2, Z: 0},
		},
		{
			name:     "90 degrees polar and azimuthal",
			angle:    blockworld.Angle3{Theta: 90, Phi: 90},
			expected: blockworld.Vec3{X: 0, Y: 1, Z: 0},
		},
		{
			name:     "90 degrees polar and 45 degrees azimuthal",
			angle:    blockworld.Angle3{Theta: 90, Phi: 45},
			expected: blockworld.Vec3{X: math.Sqrt2 / 2, Y: math.Sqrt(2) / 2, Z: 0},
		},
		{
			name:     "45 degrees polar and azimuthal",
			angle:    blockworld.Angle3{Theta: 45, Phi: 45},
			expected: blockworld.Vec3{X: 0.5, Y: 0.5, Z: math.Sqrt2 / 2},
		},
		{
			name:     "180 degrees polar",
			angle:    blockworld.Angle3{Theta: 180, Phi: 0},
			expected: blockworld.Vec3{X: 0, Y: 0, Z: -1},
		},
		{
			name:     "270 degrees polar",
			angle:    blockworld.Angle3{Theta: 270, Phi: 0},
			expected: blockworld.Vec3{X: -1, Y: 0, Z: 0},
		},
		{
			name:     "360 degrees polar",
			angle:    blockworld.Angle3{Theta: 360, Phi: 0},
			expected: blockworld.Vec3{X: 0, Y: 0, Z: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.angle.ToCartesianVec3(1)
			if !almostEqual(result, tt.expected) {
				t.Errorf("ToCartesianVec3() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestVec3_Rotate(t *testing.T) {
	tests := []struct {
		name     string
		v        blockworld.Vec3
		x, y, z  float64
		expected blockworld.Vec3
	}{
		{
			name:     "rotate 90 degrees around X axis",
			v:        blockworld.Vec3{X: 1, Y: 0, Z: 0},
			x:        90,
			y:        0,
			z:        0,
			expected: blockworld.Vec3{X: 1, Y: 0, Z: 0},
		},
		{
			name:     "rotate 90 degrees around Y axis",
			v:        blockworld.Vec3{X: 1, Y: 1, Z: 1},
			x:        0,
			y:        90,
			z:        0,
			expected: blockworld.Vec3{X: 1, Y: 1, Z: -1},
		},
		{
			name:     "rotate 180 degrees around Y axis",
			v:        blockworld.Vec3{X: 1, Y: 1, Z: 1},
			x:        0,
			y:        180,
			z:        0,
			expected: blockworld.Vec3{X: -1, Y: 1, Z: -1},
		},
		{
			name:     "rotate 90 degrees around Z axis",
			v:        blockworld.Vec3{X: 1, Y: 1, Z: 1},
			x:        0,
			y:        0,
			z:        90,
			expected: blockworld.Vec3{X: -1, Y: 1, Z: 1},
		},
		{
			name:     "rotate 180 degrees around Z axis",
			v:        blockworld.Vec3{X: 1, Y: 1, Z: 1},
			x:        0,
			y:        0,
			z:        180,
			expected: blockworld.Vec3{X: -1, Y: -1, Z: 1},
		},
		{
			name:     "rotate 360 degrees around Z axis",
			v:        blockworld.Vec3{X: 1, Y: 1, Z: 1},
			x:        0,
			y:        0,
			z:        360,
			expected: blockworld.Vec3{X: 1, Y: 1, Z: 1},
		},
		//
		{
			name:     "rotate 180 degrees around X axis",
			v:        blockworld.Vec3{X: 1, Y: 0, Z: 0},
			x:        180,
			y:        0,
			z:        0,
			expected: blockworld.Vec3{X: 1, Y: 0, Z: 0},
		},
		{
			name:     "rotate 180 degrees around Y axis",
			v:        blockworld.Vec3{X: 1, Y: 0, Z: 0},
			x:        0,
			y:        180,
			z:        0,
			expected: blockworld.Vec3{X: -1, Y: 0, Z: 0},
		},
		{
			name:     "rotate 180 degrees around Z axis",
			v:        blockworld.Vec3{X: 1, Y: 0, Z: 0},
			x:        0,
			y:        0,
			z:        180,
			expected: blockworld.Vec3{X: -1, Y: 0, Z: 0},
		},
		{
			name:     "rotate 90 degrees around Y and Z axis",
			v:        blockworld.Vec3{X: 1, Y: 1, Z: 1},
			x:        0,
			y:        90,
			z:        90,
			expected: blockworld.Vec3{X: -1, Y: 1, Z: 1},
		},
		// {
		// 	name:     "rotate 90 degrees around all axes",
		// 	v:        blockworld.Vec3{X: 1, Y: 1, Z: 1},
		// 	x:        90,
		// 	y:        90,
		// 	z:        90,
		// 	expected: blockworld.Vec3{X: -1, Y: 1, Z: -1},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.Rotate(tt.x, tt.y, tt.z)
			if !almostEqual(result, tt.expected) {
				t.Errorf("Rotate() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func almostEqual(v1, v2 blockworld.Vec3) bool {
	const epsilon = 1e-9
	return math.Abs(v1.X-v2.X) < epsilon && math.Abs(v1.Y-v2.Y) < epsilon && math.Abs(v1.Z-v2.Z) < epsilon
}

func BenchmarkMapIter(b *testing.B) {
	maps := []string{"../../maps/DragonsReach.vxl", "../../maps/AttackonDeuces.vxl"}
	for _, m := range maps {
		b.Run(m, func(b *testing.B) {
			world := blockworld.NewBlockworld()
			err := maploader.LoadMap(m, world)
			if err != nil {
				panic(err)
			}
			count := 0
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, _ = range world.Blocks() {
					count++
				}
			}
		})
	}
}

func BenchmarkViewPortRotateRay(b *testing.B) {
	viewVec := blockworld.Vec3{X: 1, Y: 0, Z: 0}
	theta := 90.0
	phi := 0.0
	rayVec := blockworld.Vec3{}
	b.Run("2-step", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rayVec = viewVec.RotateY(float64(i)).RotateZ(float64(i))
			rayVec = rayVec.RotateY(theta - 90).RotateZ(phi)
		}
	})
}

func BenchmarkImageDraw(b *testing.B) {
	var w, h = 80, 60
	var img = image.NewRGBA(image.Rect(0, 0, w, h))

	b.Run("setPixelBlack", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			img.Set(i%32, i%16, color.Black)
		}
	})
	b.Run("setPixelRed", func(b *testing.B) {
		c := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
		for i := 0; i < b.N; i++ {
			img.Set(i%32, i%16, c)
		}
	})
}

func BenchmarkWorldGetBlock(b *testing.B) {
	world := blockworld.NewBlockworld()
	world.Randomize()
	b.Run("getBlock", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var p = blockworld.Point{X: i % 512, Y: i % 512, Z: i % 64}
			world.Get(p)
		}
	})
}

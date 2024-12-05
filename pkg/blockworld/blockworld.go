package blockworld

import (
	"image/color"
	"math"
	"math/rand"
)

const blockSizePx = 25

// Point is a discrete 3D point.
type Point struct {
	X, Y, Z int
}

func (p Point) Add(p2 Point) Point {
	return Point{
		X: p.X + p2.X,
		Y: p.Y + p2.Y,
		Z: p.Z + p2.Z,
	}
}

func (p Point) Sub(p2 Point) Point {
	return Point{
		X: p.X - p2.X,
		Y: p.Y - p2.Y,
		Z: p.Z - p2.Z,
	}
}

type Vec3 struct {
	X, Y, Z float64
}

func (v Vec3) Add(v2 Vec3) Vec3 {
	return Vec3{
		X: v.X + v2.X,
		Y: v.Y + v2.Y,
		Z: v.Z + v2.Z,
	}
}

func NearestPointFromVec(pos Vec3) Point {
	return Point{
		X: int(math.Round(pos.X)),
		Y: int(math.Round(pos.Y)),
		Z: int(math.Round(pos.Z)),
	}
}

func PointFromVec(pos Vec3) Point {
	return Point{
		X: int(pos.X),
		Y: int(pos.Y),
		Z: int(pos.Z),
	}
}

type Block struct {
	Color color.Color
}

type Blockworld struct {
	blocks      map[Point]Block
	BlockSizePx int
	PlayerPos   Vec3
}

func NewBlockworld() *Blockworld {
	return &Blockworld{
		blocks:      make(map[Point]Block),
		BlockSizePx: blockSizePx,
	}
}

func (bw *Blockworld) Randomize() {
	const worldSize = 40
	colors := []color.Color{
		color.White,
		color.NRGBA{255, 0, 0, 255},   // red
		color.NRGBA{0, 255, 0, 255},   // green
		color.NRGBA{0, 0, 255, 255},   // blue
		color.NRGBA{255, 255, 0, 255}, // yellow
		color.NRGBA{255, 0, 255, 255}, // magenta
		color.NRGBA{0, 255, 255, 255}, // cyan
	}

	for x := 0; x < worldSize; x++ {
		for y := 0; y < worldSize; y++ {
			for z := 0; z < 1; z++ {
				bw.blocks[Point{x, y, z}] = Block{colors[rand.Intn(len(colors))]}
			}
		}
	}
}

func (bw *Blockworld) Get(p Point) (*Block, bool) {
	b, ok := bw.blocks[p]
	return &b, ok
}

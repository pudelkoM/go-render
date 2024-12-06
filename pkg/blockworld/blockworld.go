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

// Angle3 is a 3D angle in degrees.
type Angle3 struct {
	Theta float64 // polar, "up-down"
	Phi   float64 // azimuthal, "left-right"
}

func (a Angle3) Normalize() Angle3 {
	return Angle3{
		Theta: math.Mod(a.Theta, 360),
		Phi:   math.Mod(a.Phi, 360),
	}
}

func (a Angle3) RotatePhi(angle float64) Angle3 {
	return Angle3{
		Theta: a.Theta,
		Phi:   a.Phi + angle,
	}.Normalize()
}

func (a Angle3) ToCartesianVec3(r float64) Vec3 {
	thetaRad := a.Theta * math.Pi / 180
	phiRad := a.Phi * math.Pi / 180
	return Vec3{
		X: r * math.Sin(thetaRad) * math.Cos(phiRad),
		Y: r * math.Sin(thetaRad) * math.Sin(phiRad),
		Z: r * math.Cos(thetaRad),
	}
}

func (v Vec3) Add(v2 Vec3) Vec3 {
	return Vec3{
		X: v.X + v2.X,
		Y: v.Y + v2.Y,
		Z: v.Z + v2.Z,
	}
}

func (v Vec3) Sub(v2 Vec3) Vec3 {
	return Vec3{
		X: v.X - v2.X,
		Y: v.Y - v2.Y,
		Z: v.Z - v2.Z,
	}
}

func (v Vec3) Mul(s float64) Vec3 {
	return Vec3{
		X: v.X * s,
		Y: v.Y * s,
		Z: v.Z * s,
	}
}

// RotateX rotates the vector around the X axis.
// Positive angle is counter-clockwise. Angle is in degrees.
func (v Vec3) RotateX(angle float64) Vec3 {
	rad := angle * math.Pi / 180
	sin, cos := math.Sincos(rad)
	return Vec3{
		X: v.X,
		Y: v.Y*cos - v.Z*sin,
		Z: v.Y*sin + v.Z*cos,
	}
}

// RotateY rotates the vector around the Y axis.
// Positive angle is counter-clockwise. Angle is in degrees.
func (v Vec3) RotateY(angle float64) Vec3 {
	rad := angle * math.Pi / 180
	sin, cos := math.Sincos(rad)
	return Vec3{
		X: v.X*cos + v.Z*sin,
		Y: v.Y,
		Z: -v.X*sin + v.Z*cos,
	}
}

// RotateZ rotates the vector around the Z axis.
// Positive angle is counter-clockwise. Angle is in degrees.
func (v Vec3) RotateZ(angle float64) Vec3 {
	rad := angle * math.Pi / 180
	sin, cos := math.Sincos(rad)
	return Vec3{
		X: v.X*cos - v.Y*sin,
		Y: v.X*sin + v.Y*cos,
		Z: v.Z,
	}
}

func (v Vec3) Normalize() Vec3 {
	mag := math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
	return Vec3{
		X: v.X / mag,
		Y: v.Y / mag,
		Z: v.Z / mag,
	}
}

func (v Vec3) ToNearestPoint() Point {
	return Point{
		X: int(math.Round(v.X)),
		Y: int(math.Round(v.Y)),
		Z: int(math.Round(v.Z)),
	}
}

func (v Vec3) ToPointTrunc() Point {
	return Point{
		X: int(v.X),
		Y: int(v.Y),
		Z: int(v.Z),
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
	PlayerDir   Angle3
}

func NewBlockworld() *Blockworld {
	return &Blockworld{
		blocks:      make(map[Point]Block),
		BlockSizePx: blockSizePx,
		PlayerPos:   Vec3{X: 100, Y: 100, Z: 64},
		PlayerDir:   Angle3{Theta: 90, Phi: 45},
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

	// Generate a wall of blocks, 5 high, 1 wide, 10 long
	for x := 4; x < 5; x++ {
		for y := -5; y < 5; y++ {
			for z := 0; z < 5; z++ {
				bw.blocks[Point{x, y, z}] = Block{colors[rand.Intn(len(colors))]}
			}
		}
	}
}

func (bw *Blockworld) Get(p Point) (*Block, bool) {
	b, ok := bw.blocks[p]
	return &b, ok
}

func (bw *Blockworld) Set(x, y, z int, b Block) {
	bw.blocks[Point{x, y, z}] = b
}

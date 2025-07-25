package blockworld

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"
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

// Angle3 is a 3D angle in degrees.
type Angle3 struct {
	Theta float64 // polar, "up-down"
	Phi   float64 // azimuthal, "left-right"
}

func (a Angle3) ClampToView() Angle3 {
	if a.Theta > 180 {
		a.Theta = 180
	}
	if a.Theta < 0 {
		a.Theta = 0
	}
	return a
}

func (a Angle3) Normalize() Angle3 {
	// Normalized theta angles:
	// 181 -> 179
	// 360 -> 0
	// 359 -> 1
	// 270 -> 90
	for a.Theta > 180 {
		a.Theta = 180 - math.Mod(a.Theta, 180)
		a.Phi += 180
	}
	for a.Theta < 0 {
		a.Theta = 180 + math.Mod(a.Theta, 180)
		a.Phi += 180
	}
	for a.Phi > 360 {
		a.Phi -= 360
	}
	if a.Phi < 0 {
		a.Phi += 360
	}
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

func (a Angle3) RotateTheta(angle float64) Angle3 {
	return Angle3{
		Theta: a.Theta + angle,
		Phi:   a.Phi,
	}.Normalize()
}

func (a Angle3) ResetTheta() Angle3 {
	return Angle3{
		Theta: 90,
		Phi:   a.Phi,
	}
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

type Vec3 struct {
	X, Y, Z float64
}

func (v Vec3) String() string {
	return fmt.Sprintf("{X: %0.3f, Y: %0.3f, Z: %0.3f}", v.X, v.Y, v.Z)
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

func (v Vec3) Rotate(x, y, z float64) Vec3 {
	xRad := x * math.Pi / 180
	yRad := y * math.Pi / 180
	zRad := z * math.Pi / 180

	sinX, cosX := math.Sincos(xRad)
	sinY, cosY := math.Sincos(yRad)
	sinZ, cosZ := math.Sincos(zRad)

	return Vec3{
		X: v.X*cosY*cosZ - v.Y*cosY*sinZ + v.Z*sinY,
		Y: v.X*(sinX*sinY*cosZ+cosX*sinZ) - v.Y*(sinX*sinY*sinZ-cosX*cosZ) - v.Z*sinX*cosY,
		Z: v.X*(-cosX*sinY*cosZ-sinX*sinZ) + v.Y*(cosX*sinY*sinZ+sinX*cosZ) + v.Z*cosX*cosY,
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

func (v Vec3) Clamp(min, max float64) Vec3 {
	return Vec3{
		X: math.Max(min, math.Min(max, v.X)),
		Y: math.Max(min, math.Min(max, v.Y)),
		Z: math.Max(min, math.Min(max, v.Z)),
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

func SignAsFloat(f float64) float64 {
	return math.Float64frombits(math.Float64bits(f)&(1<<63) | 0x3FF0000000000000)
}

func (v Vec3) AdvanceToNextBlockBoundary(dir Vec3) Vec3 {
	fx := 100.
	if dir.X > 0 {
		dx := math.Floor(v.X+1) - v.X
		fx = dx / dir.X
	} else if dir.X < 0 {
		dx := math.Ceil(v.X-1) - v.X
		fx = dx / dir.X
	}
	fy := 100.
	if dir.Y > 0 {
		dy := math.Floor(v.Y+1) - v.Y
		fy = dy / dir.Y
	} else if dir.Y < 0 {
		dy := math.Ceil(v.Y-1) - v.Y
		fy = dy / dir.Y
	}
	fz := 100.
	if dir.Z > 0 {
		dz := math.Floor(v.Z+1) - v.Z
		fz = dz / dir.Z
	} else if dir.Z < 0 {
		dz := math.Ceil(v.Z-1) - v.Z
		fz = dz / dir.Z
	}

	f := 1.0001

	if fx < fy && fx < fz {
		return v.Add(dir.Mul(fx * f))
	} else if fy < fx && fy < fz {
		return v.Add(dir.Mul(fy * f))
	} else if fz < fx && fz < fy {
		return v.Add(dir.Mul(fz * f))
	} else {
		return v.Add(dir)
	}
}

type Block struct {
	Color                  color.NRGBA
	IsSet                  bool // a zero-value block is not set, i.e. air
	Reflective             bool
	DistanceToNearestBlock int16
}

type Blockworld struct {
	blocks      []Block
	svt         *SVT
	x, y, z     int
	BlockSizePx int
	PlayerPos   Vec3
	PlayerDir   Angle3
}

func NewBlockworld() *Blockworld {
	return &Blockworld{
		blocks:      make([]Block, 0),
		BlockSizePx: blockSizePx,
		PlayerPos:   Vec3{X: 170, Y: 170, Z: 64},
		PlayerDir:   Angle3{Theta: 90, Phi: 45},
	}
}

func (bw *Blockworld) Finalize() {
	t0 := time.Now()
	bw.svt.Reconcile()
	fmt.Printf("SVT reconcile took %v\n", time.Since(t0))
}

func (bw *Blockworld) RandomFill(fillPerc float64) {
	bw.svt.RandomFill(fillPerc)
	// for i := range bw.blocks {
	// 	if rand.Float64() < fillPerc {
	// 		bw.blocks[i] = Block{
	// 			Color: color.NRGBA{
	// 				R: uint8(rand.Intn(256)),
	// 				G: uint8(rand.Intn(256)),
	// 				B: uint8(rand.Intn(256)),
	// 				A: 255,
	// 			},
	// 			IsSet: true,
	// 		}
	// 	} else {
	// 		bw.blocks[i] = Block{}
	// 	}
	// }
}

func (bw *Blockworld) Randomize() {
	const worldSize = 40
	colors := []color.NRGBA{
		color.NRGBA{0, 0, 0, 0},       // white
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
				bw.Set(x, y, z, Block{Color: colors[rand.Intn(len(colors))]})
			}
		}
	}
}

func (bw *Blockworld) SetSize(x, y, z int) {
	bw.x = x
	bw.y = y
	bw.z = z
	bw.blocks = make([]Block, x*y*z)
	bw.svt = NewSVT(x, y, z)
}

func (bw *Blockworld) Blocks() []Block {
	return bw.blocks
}

func (bw *Blockworld) Get(p Point) (*Block, bool) {
	// return bw.svt.Get(p.X, p.Y, p.Z)
	// return bw.svt.GetArr(p.X, p.Y, p.Z)
	// return bw.svt.GetMapToBigNode(p.X, p.Y, p.Z)
	// return bw.svt.GetArrBigNode(p.X, p.Y, p.Z)
	// return bw.svt.GetWithCache(p.X, p.Y, p.Z)

	if (p.X < 0 || p.X >= bw.x) || (p.Y < 0 || p.Y >= bw.y) || (p.Z < 0 || p.Z >= bw.z) {
		return nil, false
	}
	b := &bw.blocks[p.X+p.Y*bw.x+p.Z*bw.x*bw.y]
	return b, b.IsSet
}

func (bw *Blockworld) GetFlatArray(x, y, z int) (*Block, bool) {
	if (x < 0 || x >= bw.x) || (y < 0 || y >= bw.y) || (z < 0 || z >= bw.z) {
		return nil, false
	}
	b := &bw.blocks[x+y*bw.x+z*bw.x*bw.y]
	return b, b.IsSet
}

func (bw *Blockworld) GetRaw(x, y, z int) (*Block, bool) {
	// return bw.svt.Get(x, y, z)
	// return bw.svt.GetArr(x, y, z)
	// return bw.svt.GetMapToBigNode(x, y, z)
	// return bw.svt.GetArrBigNode(x, y, z)
	// return bw.svt.GetWithCache(x, y, z)

	if (x < 0 || x >= bw.x) || (y < 0 || y >= bw.y) || (z < 0 || z >= bw.z) {
		return nil, false
	}
	b := &bw.blocks[x+y*bw.x+z*bw.x*bw.y]
	return b, b.IsSet
}

func (bw *Blockworld) Set(x, y, z int, b Block) {
	if (x < 0 || x >= bw.x) || (y < 0 || y >= bw.y) || (z < 0 || z >= bw.z) {
		return
	}
	b.IsSet = true
	bw.blocks[x+y*bw.x+z*bw.x*bw.y] = b

	b.IsSet = true
	bw.svt.Set(x, y, z, b)
	bw.svt.SetArr(x, y, z, b)
	bw.svt.SetMapToBigNode(x, y, z, b)
}

func (bw *Blockworld) GetJustRangeCheck(x, y, z int) (*Block, bool) {
	if x < 0 || x >= bw.x || y < 0 || y >= bw.y || z < 0 || z >= bw.z {
		return nil, false
	}
	return nil, true
}

func (bw *Blockworld) IndexCalc(x, y, z int) int {
	return x + y*bw.x + z*bw.x*bw.y
}

func (bw *Blockworld) Noop(x, y, z int) int {
	return 0
}

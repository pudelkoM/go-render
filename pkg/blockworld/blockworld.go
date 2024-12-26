package blockworld

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/pudelkoM/go-render/pkg/utils"
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

func (p Point) ToVec3() Vec3 {
	return Vec3{
		X: float64(p.X),
		Y: float64(p.Y),
		Z: float64(p.Z),
	}
}

func (p Point) String() string {
	return fmt.Sprintf("{X: %d, Y: %d, Z: %d}", p.X, p.Y, p.Z)
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

func (v Vec3) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
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

const (
	BLOCK_FACE_FRONT = iota
	BLOCK_FACE_BACK
	BLOCK_FACE_LEFT
	BLOCK_FACE_RIGHT
	BLOCK_FACE_TOP
	BLOCK_FACE_BOTTOM
)

func epsilonEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-3
}

func GetBlockFace(pos Vec3, block Point) int {
	if epsilonEqual(pos.Z, float64(block.Z)) {
		return BLOCK_FACE_BOTTOM
	}
	if epsilonEqual(pos.Z, float64(block.Z+1)) {
		return BLOCK_FACE_TOP
	}
	if epsilonEqual(pos.X, float64(block.X)) {
		return BLOCK_FACE_LEFT
	}
	if epsilonEqual(pos.X, float64(block.X+1)) {
		return BLOCK_FACE_RIGHT
	}
	if epsilonEqual(pos.Y, float64(block.Y)) {
		return BLOCK_FACE_BACK
	}
	if epsilonEqual(pos.Y, float64(block.Y+1)) {
		return BLOCK_FACE_FRONT
	}

	return -1
}

type Block struct {
	Color                  color.NRGBA
	IsSet                  bool // a zero-value block is not set, i.e. air
	Reflective             bool
	DistanceToNearestBlock int16
	IsLightSource          bool
}

type Blockworld struct {
	blocks      []Block
	x, y, z     int
	BlockSizePx int
	PlayerPos   Vec3
	PlayerDir   Angle3
	Lights      []Point
}

func NewBlockworld() *Blockworld {
	return &Blockworld{
		blocks:      make([]Block, 0),
		BlockSizePx: blockSizePx,
		PlayerPos:   Vec3{X: 170, Y: 170, Z: 64},
		PlayerDir:   Angle3{Theta: 90, Phi: 45},
	}
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
	bw.Lights = make([]Point, 0)
}

func (bw *Blockworld) Blocks() []Block {
	return bw.blocks
}

func (bw *Blockworld) Get(p Point) (*Block, bool) {
	if (p.X < 0 || p.X >= bw.x) || (p.Y < 0 || p.Y >= bw.y) || (p.Z < 0 || p.Z >= bw.z) {
		return nil, false
	}
	b := &bw.blocks[p.X+p.Y*bw.x+p.Z*bw.x*bw.y]
	return b, b.IsSet
}

func (bw *Blockworld) GetRaw(x, y, z int) (*Block, bool) {
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
}

func (bw *Blockworld) CreateLightBlock(x, y, z int) {
	b, _ := bw.GetRaw(x, y, z)
	b.IsLightSource = true
	b.IsSet = true
	b.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	bw.Lights = append(bw.Lights, Point{X: x, Y: y, Z: z})
}

func (bw *Blockworld) RayMarchSdf(start, dir Vec3) Vec3 {
	// skipBlocks := []int16{}
	// totalSkippedBlocks := 0

	for {
		bc, _ := bw.Get(start.ToPointTrunc())
		if bc == nil || bc.IsSet || bc.IsLightSource {
			break
		}
		if bc.DistanceToNearestBlock <= 3 {
			start = start.AdvanceToNextBlockBoundary(dir)
			continue
		}
		const sqrt3 = 1.73205080757
		start = start.Add(dir.Mul(float64(bc.DistanceToNearestBlock-2) / sqrt3))
		// totalSkippedBlocks += int(bc.DistanceToNearestBlock - 2)
		// skipBlocks = append(skipBlocks, bc.DistanceToNearestBlock)
	}

	// if x == img.Rect.Dx()/2 && y == img.Rect.Dy()/2 {
	// 	fmt.Println("skip blocks", skipBlocks, "total skipped", totalSkippedBlocks)
	// }
	return start
}

func (bw *Blockworld) ComputeShadows() {
	if len(bw.Lights) != 1 {
		return
	}
	l := bw.Lights[0].ToVec3()
	l = l.Add(Vec3{X: 0.5, Y: 0.5, Z: 0.5})

	// lb, _ := bw.Get(l)
	// lb.IsSet = false
	// defer func() { lb.IsSet = true }()

	for z := 0; z < bw.z; z++ {
		for y := 0; y < bw.y; y++ {
			for x := 0; x < bw.x; x++ {
				b, _ := bw.GetRaw(x, y, z)
				if !b.IsSet || b.IsLightSource {
					continue
				}

				b.IsSet = false

				c := Point{X: x, Y: y, Z: z}.ToVec3().Add(Vec3{X: 0.5, Y: 0.5, Z: 1})
				// rayVec := l.ToVec3().Add(Vec3{X: 0.5, Y: 0.5, Z: 0.5}).Sub(c).Normalize()
				// c := Point{X: x, Y: y, Z: z}.ToVec3()
				rayVec := l.Sub(c).Normalize()
				newPos := bw.RayMarchSdf(c, rayVec)

				b2, _ := bw.Get(newPos.ToPointTrunc())
				if b2 != nil && b2.IsLightSource {
					// b.Color = color.NRGBA{R: 200, G: 0, B: 0, A: 255}
				} else {
					// Block is in shadow
					b.Color = utils.ColorDarken(b.Color, 0.3)
					// b.Color = color.NRGBA{R: 200, G: 0, B: 0, A: 255}
				}

				// if x == 247 && y == 256 && z == 27 {
				// 	b.Color = color.NRGBA{R: 0, G: 0, B: 255, A: 255}
				// 	fmt.Println("c", c)
				// 	fmt.Println("rayVec", rayVec)
				// 	fmt.Println("newPos", newPos.ToPointTrunc())
				// 	fmt.Println("b2", b2)
				// }

				b.IsSet = true
			}
		}
	}

}

func (world *Blockworld) ComputeNearestBlocks() {
	// Fill X-Axis values.
	for z := 0; z < world.z; z++ {
		for y := 0; y < world.y; y++ {
			d := int16(512)
			for x := 0; x < world.x; x++ {
				d++
				b, _ := world.GetRaw(x, y, z)
				if b.IsSet {
					d = 0
					continue
				}
				if b.DistanceToNearestBlock == 0 { // not visited before
					b.DistanceToNearestBlock = d
				} else {
					d = min(b.DistanceToNearestBlock, d)
					b.DistanceToNearestBlock = d
				}
			}
			d = 512
			for x := world.x - 1; x >= 0; x-- {
				d++
				b, _ := world.GetRaw(x, y, z)
				if b.IsSet {
					d = 0
					continue
				}
				if b.DistanceToNearestBlock == 0 { // not visited before
					b.DistanceToNearestBlock = d
				} else {
					d = min(b.DistanceToNearestBlock, d)
					b.DistanceToNearestBlock = d
				}
			}
		}
	}

	// Fill Y-Axis values.
	for z := 0; z < world.z; z++ {
		for x := 0; x < world.x; x++ {
			d := int16(512)
			for y := 0; y < world.y; y++ {
				d++
				b, _ := world.GetRaw(x, y, z)
				if b.IsSet {
					d = 0
					continue
				}
				if b.DistanceToNearestBlock == 0 { // not visited before
					b.DistanceToNearestBlock = d
				} else {
					d = min(b.DistanceToNearestBlock, d)
					b.DistanceToNearestBlock = d
				}
			}
			d = 512
			for y := world.y - 1; y >= 0; y-- {
				d++
				b, _ := world.GetRaw(x, y, z)
				if b.IsSet {
					d = 0
					continue
				}
				if b.DistanceToNearestBlock == 0 { // not visited before
					b.DistanceToNearestBlock = d
				} else {
					d = min(b.DistanceToNearestBlock, d)
					b.DistanceToNearestBlock = d
				}
			}
		}
	}

	// Fill Z-Axis values.
	for x := 0; x < world.x; x++ {
		for y := 0; y < world.y; y++ {
			d := int16(512)
			for z := 0; z < world.z; z++ {
				d++
				b, _ := world.GetRaw(x, y, z)
				if b.IsSet {
					d = 0
					continue
				}
				if b.DistanceToNearestBlock == 0 { // not visited before
					b.DistanceToNearestBlock = d
				} else {
					d = min(b.DistanceToNearestBlock, d)
					b.DistanceToNearestBlock = d
				}
			}
			d = 512
			for z := world.z - 1; z >= 0; z-- {
				d++
				b, _ := world.GetRaw(x, y, z)
				if b.IsSet {
					d = 0
					continue
				}
				if b.DistanceToNearestBlock == 0 { // not visited before
					b.DistanceToNearestBlock = d
				} else {
					d = min(b.DistanceToNearestBlock, d)
					b.DistanceToNearestBlock = d
				}
			}
		}
	}
}

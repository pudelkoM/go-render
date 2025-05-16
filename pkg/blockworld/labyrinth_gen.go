package blockworld

import (
	"fmt"
	"image/color"
	"math/rand"
)

const (
	floor_z = 3
)

// GenerateLabyrinth creates a 2D labyrinth using recursive backtracking
func GenerateLabyrinth(world *Blockworld, x, y int) {
	world.SetSize(x*2, y*2, 64)

	world.PlayerPos = Vec3{X: float64(x) / 2, Y: float64(y / 2), Z: 33}
	world.PlayerDir = Angle3{Theta: 95, Phi: 325}
	world.PlayerFovHDeg = 90

	grayWithVariation := func(base, vari int) color.NRGBA {
		rInt := rand.Intn(vari)
		return color.NRGBA{
			R: uint8(base + rInt),
			G: uint8(base + rInt),
			B: uint8(base + rInt),
			A: 255,
		}
	}

	// Initialize the grid with walls
	for y := range y {
		for x := range x {
			// Floor
			world.Set(x, y, floor_z-1, Block{
				Color: grayWithVariation(40, 10),
				IsSet: true,
			})

			// // Ceiling
			// world.Set(x, y, floor_z+1, Block{
			// 	Color: grayWithVariation(40, 10),
			// 	IsSet: true,
			// })

			// Walls
			world.Set(x, y, floor_z, Block{
				Color: grayWithVariation(100, 50),
				IsSet: true,
			})
		}
	}

	// Start carving paths from a random starting point
	startX, startY := rand.Intn(x/2)*2+1, rand.Intn(y/2)*2+1
	carvePath(world, startX, startY)

	// Set goal point
	world.Set(startX, startY, floor_z-1, Block{
		Color: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
		IsSet: true,
	})
	fmt.Printf("Goal point: (%d, %d)\n", startX, startY)

	// Find random spawn point
	for {
		spawnX, spawnY := rand.Intn(x), rand.Intn(y)
		_, isSet := world.GetRaw(spawnX, spawnY, floor_z)
		if !isSet {
			world.Set(spawnX, spawnY, floor_z-1, Block{
				Color: color.NRGBA{R: 0, G: 255, B: 0, A: 255},
				IsSet: true,
			})

			fmt.Printf("Spawn point: (%d, %d)\n", spawnX, spawnY)
			// TODO: set player pos to startX, startY
			world.PlayerPos = Vec3{X: float64(spawnX) + 0.5, Y: float64(spawnY) + 0.5, Z: floor_z + 0.5}
			world.PlayerDir = Angle3{Theta: 90, Phi: 0}
			break
		}
	}

	// world.CreateLightBlock(x/2, y/2, 60)

	world.Finalize()
}

// carvePath recursively carves paths in the grid
func carvePath(world *Blockworld, x, y int) {
	maxX, maxY, _ := world.GetSize()
	directions := []struct{ dx, dy int }{
		{dx: 0, dy: -2}, // Up
		{dx: 2, dy: 0},  // Right
		{dx: 0, dy: 2},  // Down
		{dx: -2, dy: 0}, // Left
	}

	// Shuffle directions to create randomness
	rand.Shuffle(len(directions), func(i, j int) {
		directions[i], directions[j] = directions[j], directions[i]
	})

	// Mark the current cell as a path
	world.Set(x, y, floor_z, Block{
		IsSet: false,
	})

	// Explore each direction
	for _, dir := range directions {
		nx, ny := x+dir.dx, y+dir.dy
		// Check if the next cell is within bounds and is a wall
		_, isSet := world.GetRaw(nx, ny, floor_z)
		if ny > 0 && ny < maxY-1 && nx > 0 && nx < maxX-1 && isSet {
			// Carve a path to the next cell
			world.Set(x+dir.dx/2, y+dir.dy/2, floor_z, Block{
				IsSet: false,
			})
			carvePath(world, nx, ny)
		}
	}
}

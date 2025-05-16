package labsolver

import (
	"fmt"
	"math"

	"github.com/pudelkoM/go-render/pkg/blockworld"
	"github.com/pudelkoM/go-render/pkg/utils"
)

func blockIsGoal(b blockworld.Block) bool {
	return b.IsSet && b.Color.R == 255 && b.Color.G == 0 && b.Color.B == 0
}

func blockIsWall(b blockworld.Block) bool {
	return b.IsSet && !blockIsGoal(b)
}

func blockIsPath(b blockworld.Block) bool {
	return !b.IsSet
}

func isAt90Degrees(angle float64) bool {
	return utils.AlmostEqual(angle, 0) ||
		utils.AlmostEqual(angle, 90) ||
		utils.AlmostEqual(angle, 180) ||
		utils.AlmostEqual(angle, 270)
}

const (
	STOPPED          = iota // Stopped means the player has completed a movement and can plan the next one
	MOVING_FORWARD          // Moving means the player is currently moving forward
	TURNING_LEFT            // Turning left means the player is currently turning left
	TURNING_RIGHT           // Turning right means the player is currently turning right
	TURNING_AROUND          // Turning around means the player is currently turning around
	TURNING_AROUND_2        // Turning around 2 means the player is currently turning around 2
	DONE                    // Done means the player has completed the labyrinth
)

// LabSolver is a structure that contains the state of the labyrinth solver.
type LabSolver struct {
	state int
}

func (state *LabSolver) String() string {
	switch state.state {
	case STOPPED:
		return "STOPPED"
	case MOVING_FORWARD:
		return "MOVING_FORWARD"
	case TURNING_LEFT:
		return "TURNING_LEFT"
	case TURNING_RIGHT:
		return "TURNING_RIGHT"
	case TURNING_AROUND:
		return "TURNING_AROUND"
	case TURNING_AROUND_2:
		return "TURNING_AROUND_2"
	case DONE:
		return "DONE"
	default:
		return "UNKNOWN"
	}
}

// Advance gets the current world state and navigates the player towards the goal
// inside the labyrinth.
func Advance(world *blockworld.Blockworld, state *LabSolver, frameCount int64) {
	const forwardSpeed = 0.05
	const turnSpeed = 3

	if frameCount%30 != 0 {
		// return
		fmt.Printf("PlayerPos: %v\n", world.PlayerPos)
		fmt.Printf("PlayerDir: %v\n", world.PlayerDir)
		fmt.Printf("State: %v\n", state)
	}

	if state.state == MOVING_FORWARD {
		_, fracX := math.Modf(world.PlayerPos.X)
		_, fracY := math.Modf(world.PlayerPos.Y)
		// Check if the player is at the center of a block
		if utils.AlmostEqual(fracX, 0.5) && utils.AlmostEqual(fracY, 0.5) {
			// If the player is at the center of a block, stop moving
			state.state = STOPPED
			return
		} else {
			// If the player is not at the center of a block, keep moving
			world.PlayerPos = world.PlayerPos.Add(world.PlayerDir.ToCartesianVec3(forwardSpeed))
			return
		}
	} else if state.state == TURNING_LEFT || state.state == TURNING_RIGHT {
		// Check if the player direction is at a 90 degree angle
		if isAt90Degrees(world.PlayerDir.Phi) {
			// If the player direction is at a 90 degree angle, stop turning and move forward
			state.state = MOVING_FORWARD
			// Move the player forward
			world.PlayerPos = world.PlayerPos.Add(world.PlayerDir.ToCartesianVec3(forwardSpeed))
			return
		} else {
			// If the player direction is not at a 90 degree angle, keep turning
			if state.state == TURNING_LEFT {
				world.PlayerDir = world.PlayerDir.RotatePhi(-turnSpeed)
			} else {
				world.PlayerDir = world.PlayerDir.RotatePhi(turnSpeed)
			}
			return
		}
	} else if state.state == TURNING_AROUND {
		// Check if the player direction is at a 90 degree angle
		if isAt90Degrees(world.PlayerDir.Phi) {
			// If the player direction is at a 90 degree angle, complete the turn
			state.state = TURNING_AROUND_2
			world.PlayerDir = world.PlayerDir.RotatePhi(-turnSpeed)
			return
		} else {
			// If the player direction is not at a 90 degree angle, keep turning
			world.PlayerDir = world.PlayerDir.RotatePhi(-turnSpeed)
			return
		}
	} else if state.state == TURNING_AROUND_2 {
		// Check if the player direction is at a 90 degree angle
		if isAt90Degrees(world.PlayerDir.Phi) {
			// If the player direction is at a 90 degree angle, stop turning and move forward
			state.state = MOVING_FORWARD
			world.PlayerPos = world.PlayerPos.Add(world.PlayerDir.ToCartesianVec3(forwardSpeed))
			return
		} else {
			// If the player direction is not at a 90 degree angle, keep turning
			world.PlayerDir = world.PlayerDir.RotatePhi(-turnSpeed)
			return
		}
	} else if state.state == DONE {
		return
	}

	// Get the block in front of the player
	front, _ := world.Get(world.PlayerPos.Add(world.PlayerDir.ToCartesianVec3(1)).ToPointTrunc())
	// Get the block to the left of the player
	left, _ := world.Get(world.PlayerPos.Add(world.PlayerDir.RotatePhi(-90).ToCartesianVec3(1)).ToPointTrunc())
	// Get the block to the right of the player
	right, _ := world.Get(world.PlayerPos.Add(world.PlayerDir.RotatePhi(90).ToCartesianVec3(1)).ToPointTrunc())
	// Get the block under the player
	down, _ := world.Get(world.PlayerPos.Add(blockworld.Vec3{X: 0, Y: 0, Z: -1}).ToPointTrunc())

	// Always follow the right wall
	if blockIsGoal(*down) {
		// If it is the goal, stop the player
		state.state = DONE
		world.PlayerDir = world.PlayerDir.RotateTheta(45)
		return
	}
	if blockIsPath(*right) {
		// If the right block is a path, turn right
		state.state = TURNING_RIGHT
		world.PlayerDir = world.PlayerDir.RotatePhi(turnSpeed)
	} else if blockIsPath(*front) {
		// If the block in front is a path, move forward
		state.state = MOVING_FORWARD
		world.PlayerPos = world.PlayerPos.Add(world.PlayerDir.ToCartesianVec3(forwardSpeed))
	} else if blockIsWall(*front) {
		// If the block in front is a wall, check the left and right blocks
		if blockIsPath(*right) {
			// If the right block is a path, turn right
			state.state = TURNING_RIGHT
			world.PlayerDir = world.PlayerDir.RotatePhi(turnSpeed)
		} else if blockIsPath(*left) {
			// If the left block is a path, turn left
			state.state = TURNING_LEFT
			world.PlayerDir = world.PlayerDir.RotatePhi(-turnSpeed)
		} else {
			// If both left and right are walls, turn around
			state.state = TURNING_AROUND
			world.PlayerDir = world.PlayerDir.RotatePhi(-turnSpeed)
		}
	}
}

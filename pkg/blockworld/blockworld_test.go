package blockworld

import (
	"math"
	"testing"
)

func TestVec3_RotateZ(t *testing.T) {
	tests := []struct {
		name     string
		v        Vec3
		angle    float64
		expected Vec3
	}{
		{
			name:     "rotate 90 degrees",
			v:        Vec3{X: 1, Y: 0, Z: 0},
			angle:    90,
			expected: Vec3{X: 0, Y: 1, Z: 0},
		},
		{
			name:     "rotate 180 degrees",
			v:        Vec3{X: 1, Y: 0, Z: 0},
			angle:    180,
			expected: Vec3{X: -1, Y: 0, Z: 0},
		},
		{
			name:     "rotate 270 degrees",
			v:        Vec3{X: 1, Y: 0, Z: 0},
			angle:    270,
			expected: Vec3{X: 0, Y: -1, Z: 0},
		},
		{
			name:     "rotate 360 degrees",
			v:        Vec3{X: 1, Y: 0, Z: 0},
			angle:    360,
			expected: Vec3{X: 1, Y: 0, Z: 0},
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

func almostEqual(v1, v2 Vec3) bool {
	const epsilon = 1e-9
	return math.Abs(v1.X-v2.X) < epsilon && math.Abs(v1.Y-v2.Y) < epsilon && math.Abs(v1.Z-v2.Z) < epsilon
}

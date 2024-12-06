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

func TestAngle3_ToCartesianVec3(t *testing.T) {
	tests := []struct {
		name     string
		angle    Angle3
		expected Vec3
	}{
		{
			name:     "zero angles",
			angle:    Angle3{Theta: 0, Phi: 0},
			expected: Vec3{X: 0, Y: 0, Z: 1},
		},
		{
			name:     "90 degrees polar",
			angle:    Angle3{Theta: 90, Phi: 0},
			expected: Vec3{X: 1, Y: 0, Z: 0},
		},
		{
			name:     "45 degrees polar",
			angle:    Angle3{Theta: 45, Phi: 0},
			expected: Vec3{X: math.Sqrt2 / 2, Y: 0, Z: math.Sqrt2 / 2},
		},
		{
			name:     "90 degrees azimuthal",
			angle:    Angle3{Theta: 0, Phi: 90},
			expected: Vec3{X: 0, Y: 0, Z: 1},
		},
		{
			name:     "90 degrees polar 45 degrees azimuthal",
			angle:    Angle3{Theta: 90, Phi: 45},
			expected: Vec3{X: math.Sqrt2 / 2, Y: math.Sqrt2 / 2, Z: 0},
		},
		{
			name:     "90 degrees polar and azimuthal",
			angle:    Angle3{Theta: 90, Phi: 90},
			expected: Vec3{X: 0, Y: 1, Z: 0},
		},
		{
			name:     "90 degrees polar and 45 degrees azimuthal",
			angle:    Angle3{Theta: 90, Phi: 45},
			expected: Vec3{X: math.Sqrt2 / 2, Y: math.Sqrt(2) / 2, Z: 0},
		},
		{
			name:     "45 degrees polar and azimuthal",
			angle:    Angle3{Theta: 45, Phi: 45},
			expected: Vec3{X: 0.5, Y: 0.5, Z: math.Sqrt2 / 2},
		},
		{
			name:     "180 degrees polar",
			angle:    Angle3{Theta: 180, Phi: 0},
			expected: Vec3{X: 0, Y: 0, Z: -1},
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

func almostEqual(v1, v2 Vec3) bool {
	const epsilon = 1e-9
	return math.Abs(v1.X-v2.X) < epsilon && math.Abs(v1.Y-v2.Y) < epsilon && math.Abs(v1.Z-v2.Z) < epsilon
}

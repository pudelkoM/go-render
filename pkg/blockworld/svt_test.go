package blockworld

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFoo(t *testing.T) {
	svt := NewSVT()
	svt.RandomFill(0.1)

	require.Equal(t, 512, svt.Dx(), "Expected root leafs to have a width of n blocks")
	require.Equal(t, 512, svt.Dy(), "Expected root leafs to have a height of n blocks")
	require.Equal(t, 512, svt.Dz(), "Expected root leafs to have a depth of n blocks")

	// t.Logf("SVT structure:\n%+v\n", svt.root)
	t.Log("SVT structure:\n", svt)

	_, ok := svt.Get(0, 0, 0) // Accessing a block to ensure it works
	require.True(t, ok, "Expected to get a block at (0, 0, 0)")
}

func BenchmarkTraverseSVT(b *testing.B) {
	for _, f := range []float64{0, 0.01, 0.1, 0.3, 0.5, 0.7, 0.9, 0.99, 1} {
		b.Run(fmt.Sprintf("TraverseSVT-%v", f), func(b *testing.B) {
			svt := NewSVT()
			svt.RandomFill(f)
			dx, dy, dz := svt.Dx(), svt.Dy(), svt.Dz()
			x := rand.Intn(dx)
			y := rand.Intn(dy)
			z := rand.Intn(dz)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Access random location
				_, _ = svt.Get(x, y, z)
				x = (x + 1) % dx
				y = (y - 1) % dy
				z = (z + 2) % dz

				// for z := 0; z < svt.Dz(); z++ {
				// 	for y := 0; y < svt.Dy(); y++ {
				// 		for x := 0; x < svt.Dx(); x++ {
				// 			_, _ = svt.Get(x, y, z) // Accessing each block
				// 		}
				// 	}
				// }

				// _, _ = svt.Get(i%dx, i%dy, i%dz) // Accessing random blocks
				// _, _ = svt.Get((i+1)%dx, i%dy, (i+1)%dz)
				// _, _ = svt.Get((i+2)%dx, i%dy, i%dz)
				// _, _ = svt.Get((i+3)%dx, i%dy, (i-1)%dz)
				// _, _ = svt.Get((i+3)%dx, (i+1)%dy, i%dz)
				// _, _ = svt.Get((i+3)%dx, (i+2)%dy, (i+1)%dz)
				// _, _ = svt.Get((i+3)%dx, (i+3)%dy, i%dz)
			}
		})
	}
}

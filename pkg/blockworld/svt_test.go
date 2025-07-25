package blockworld

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	lfucache "github.com/NdoleStudio/lfu-cache/v2"
	"github.com/hashicorp/golang-lru/v2/simplelru"
	"github.com/kamstrup/intmap"
	"github.com/stretchr/testify/require"
)

func TestFoo(t *testing.T) {
	svt := NewSVT(64, 64, 64)
	svt.RandomFill(0.1)

	require.Equal(t, 64, svt.Dx(), "Expected root leafs to have a width of n blocks")
	require.Equal(t, 64, svt.Dy(), "Expected root leafs to have a height of n blocks")
	require.Equal(t, 64, svt.Dz(), "Expected root leafs to have a depth of n blocks")

	// t.Logf("SVT structure:\n%+v\n", svt.root)
	t.Log("SVT structure:\n", svt)

	_, ok := svt.Get(0, 0, 0) // Accessing a block to ensure it works
	require.True(t, ok, "Expected to get a block at (0, 0, 0)")

	_, ok = svt.GetArrBigNode(0, 0, 0) // Accessing a block to ensure it works
	require.True(t, ok, "Expected to get a block at (0, 0, 0)")
}

func TestSvtGetSet(t *testing.T) {
	svt := NewSVT(64, 64, 64)

	svt.Set(0, 0, 0, Block{IsSet: true})
	svt.SetArr(0, 0, 0, Block{IsSet: true})
	svt.Reconcile()
	_, ok := svt.Get(0, 0, 0) // Accessing a block to ensure it works
	require.True(t, ok, "Expected to get a block at (0, 0, 0)")
	_, ok = svt.GetArrBigNode(0, 0, 0) // Accessing a block to ensure it works
	require.True(t, ok, "Expected to get a block at (0, 0, 0)")

	_, ok = svt.Get(0, 1, 0) // not set
	require.False(t, ok, "Not expected to get a block at (0, 1, 0)")
	_, ok = svt.GetArrBigNode(0, 1, 0) // not set
	require.False(t, ok, "Not expected to get a block at (0, 1, 0)")
}

func BenchmarkLeafGet(b *testing.B) {
	l := leaf{}
	l.Set(32, 32, 32, Block{IsSet: true})
	l.Reconcile()

	b.Run("GetSameBlock", func(b *testing.B) {
		for b.Loop() {
			_, ok := l.GetRelative(32, 32, 32)
			if !ok {
				b.Fatal("Expected to get a block at (32, 32, 32)")
			}
		}
	})

	b.Run("GetUnsetBlock", func(b *testing.B) {
		for b.Loop() {
			_, ok := l.GetRelative(31, 31, 31)
			if ok {
				b.Fatal("Not expected to get a block at (31, 31, 31)")
			}
		}
	})

	b.Run("GetRandomBlock", func(b *testing.B) {
		x := rand.Intn(64)
		y := rand.Intn(64)
		z := rand.Intn(64)
		for b.Loop() {
			_, _ = l.GetRelative(x, y, z)
			x = (x + 1)
			y = (y + 2)
			z = (z + 4)
		}
	})
}

func BenchmarkNode0Get(b *testing.B) {
	n := node0{}
	n.Set(32, 32, 32, Block{IsSet: true})
	n.Reconcile()

	b.Run("GetSameBlock", func(b *testing.B) {
		for b.Loop() {
			_, ok := n.GetRelative(32, 32, 32)
			if !ok {
				b.Fatal("Expected to get a block at (32, 32, 32)")
			}
		}
	})

	b.Run("GetSameBlockInline", func(b *testing.B) {
		for b.Loop() {
			_, ok := n.GetInline(32, 32, 32)
			if !ok {
				b.Fatal("Expected to get a block at (32, 32, 32)")
			}
		}
	})

	b.Run("GetUnsetBlock", func(b *testing.B) {
		for b.Loop() {
			_, ok := n.GetRelative(31, 31, 31)
			if ok {
				b.Fatal("Not expected to get a block at (31, 31, 31)")
			}
		}
	})

	b.Run("GetUnsetBlockInline", func(b *testing.B) {
		for b.Loop() {
			_, ok := n.GetInline(31, 31, 31)
			if ok {
				b.Fatal("Not expected to get a block at (31, 31, 31)")
			}
		}
	})

	b.Run("GetRandomBlock", func(b *testing.B) {
		x := rand.Intn(64)
		y := rand.Intn(64)
		z := rand.Intn(64)
		for b.Loop() {
			_, _ = n.GetRelative(x, y, z)
			x = (x + 1)
			y = (y + 2)
			z = (z + 4)
		}
	})

	b.Run("GetRandomBlockInline", func(b *testing.B) {
		x := rand.Intn(64)
		y := rand.Intn(64)
		z := rand.Intn(64)
		for b.Loop() {
			_, _ = n.GetInline(x, y, z)
			x = (x + 1)
			y = (y + 2)
			z = (z + 4)
		}
	})
}

const cacheSize = 64

func BenchmarkSimpleLruCache(b *testing.B) {
	b.Run("GetEmpty", func(b *testing.B) {
		l, _ := simplelru.NewLRU[uint64, *node0](cacheSize, nil)
		b.ResetTimer()
		for range b.N {
			l.Get(rand.Uint64())
		}
	})

	for _, f := range []float64{0.01, 0.1, 0.3, 0.5, 0.7, 0.9, 0.99} {
		const maxN = 1024 * 16
		b.Run(fmt.Sprintf("Get-%v", f), func(b *testing.B) {
			l, _ := simplelru.NewLRU[uint64, *node0](cacheSize, nil)
			for i := range uint64(maxN) {
				if rand.Float64() < f {
					l.Add(i, nil)
				}
			}
			b.ResetTimer()
			for range b.N {
				l.Get(uint64(rand.Int63n(maxN)))
			}
		})
	}
}

func BenchmarkLfuCache(b *testing.B) {
	b.Run("GetEmpty", func(b *testing.B) {
		l, _ := lfucache.New[uint64, *node0](cacheSize)
		b.ResetTimer()
		for range b.N {
			l.Get(rand.Uint64())
		}
	})

	for _, f := range []float64{0.01, 0.1, 0.3, 0.5, 0.7, 0.9, 0.99} {
		const maxN = 1024 * 16
		b.Run(fmt.Sprintf("Get-%v", f), func(b *testing.B) {
			l, _ := lfucache.New[uint64, *node0](cacheSize)
			for i := range uint64(maxN) {
				if rand.Float64() < f {
					l.Set(i, nil)
				}
			}
			b.ResetTimer()
			for range b.N {
				l.Get(uint64(rand.Int63n(maxN)))
			}
		})
	}
}

// Simple linear congruential generator (LCG)
func uintPRNG(seed uint64) func() uint64 {
	return func() uint64 {
		seed = (6364136223846793005*seed + 1) & 0xFFFFFFFFFFFFFFFF
		return seed
	}
}

func BenchmarkMap(b *testing.B) {
	const initSize = 1024 * 16
	const maxN = 1 << 20

	b.Run("RandRef", func(b *testing.B) {
		for b.Loop() {
			_ = uint64(rand.Int63n(maxN))
		}
	})
	b.Run("MyRand", func(b *testing.B) {
		f := uintPRNG(uint64(time.Now().Unix()))
		for b.Loop() {
			_ = f()
		}
	})
	for _, f := range []float64{0, 0.01, 0.1, 0.3, 0.5, 0.7, 0.9, 0.99, 1} {
		seed := uint64(time.Now().Unix())
		r := uintPRNG(seed)
		b.Run(fmt.Sprintf("StdGet-%v", f), func(b *testing.B) {
			m := make(map[uint64]*node0, initSize)
			for i := range uint64(maxN) {
				if rand.Float64() < f {
					m[i] = nil
				}
			}
			for b.Loop() {
				// _, _ = m[uint64(rand.Int63n(maxN))]
				_, _ = m[r()]
			}
		})
		r = uintPRNG(seed)
		b.Run(fmt.Sprintf("IntGet-%v", f), func(b *testing.B) {
			m := intmap.New[uint64, *node0](initSize)
			for i := range uint64(maxN) {
				if rand.Float64() < f {
					m.Put(i, nil)
				}
			}
			for b.Loop() {
				// m.Get(uint64(rand.Int63n(maxN)))
				m.Get(r())
			}
		})
	}
}

func BenchmarkTraverseSVT(b *testing.B) {
	// for _, f := range []float64{0, 0.01, 0.1, 0.3, 0.5, 0.7, 0.9, 0.99, 1} {
	// for _, f := range []float64{0.01, 0.5, 0.99} {
	for _, s := range []int{16, 512, 1024} {
		// for _, s := range []int{128} {
		for _, f := range []float64{0.01} {
			svt := NewSVT(s, s, s)
			svt.RandomFill(f)
			b.Run(fmt.Sprintf("GetJustRangeCheck-%v-%v", s, f), func(b *testing.B) {
				dx, dy, dz := svt.Dx(), svt.Dy(), svt.Dz()
				x := rand.Intn(dx)
				y := rand.Intn(dy)
				z := rand.Intn(dz)
				b.ResetTimer()
				for b.Loop() {
					_, _ = svt.GetJustRangeCheck(x, y, z)
					x = (x + 1) % dx
					y = (y + 2) % dy
					z = (z + 4) % dz
				}
			})
			b.Run(fmt.Sprintf("Get-%v-%v", s, f), func(b *testing.B) {
				dx, dy, dz := svt.Dx(), svt.Dy(), svt.Dz()
				x := rand.Intn(dx)
				y := rand.Intn(dy)
				z := rand.Intn(dz)
				b.ResetTimer()
				for b.Loop() {
					_, _ = svt.Get(x, y, z)
					x = (x + 1) % dx
					y = (y + 2) % dy
					z = (z + 4) % dz
				}
			})
			b.Run(fmt.Sprintf("GetInline-%v-%v", s, f), func(b *testing.B) {
				dx, dy, dz := svt.Dx(), svt.Dy(), svt.Dz()
				x := rand.Intn(dx)
				y := rand.Intn(dy)
				z := rand.Intn(dz)
				b.ResetTimer()
				for b.Loop() {
					_, _ = svt.GetInline(x, y, z)
					x = (x + 1) % dx
					y = (y + 2) % dy
					z = (z + 4) % dz
				}
			})
			b.Run(fmt.Sprintf("GetArrBigNode-%v-%v", s, f), func(b *testing.B) {
				dx, dy, dz := svt.Dx(), svt.Dy(), svt.Dz()
				x := rand.Intn(dx)
				y := rand.Intn(dy)
				z := rand.Intn(dz)
				b.ResetTimer()
				for b.Loop() {
					_, _ = svt.GetArrBigNode(x, y, z)
					x = (x + 1) % dx
					y = (y + 2) % dy
					z = (z + 4) % dz
				}
			})
			b.Run(fmt.Sprintf("GetMapToBigNode-%v-%v", s, f), func(b *testing.B) {
				dx, dy, dz := svt.Dx(), svt.Dy(), svt.Dz()
				x := rand.Intn(dx)
				y := rand.Intn(dy)
				z := rand.Intn(dz)
				b.ResetTimer()
				for b.Loop() {
					_, _ = svt.GetMapToBigNode(x, y, z)
					x = (x + 1) % dx
					y = (y + 2) % dy
					z = (z + 4) % dz
				}
			})
			b.Run(fmt.Sprintf("GetMapToPoolBigNode-%v-%v", s, f), func(b *testing.B) {
				b.Logf("pool size: %d", len(svt.bigNodePool))
				b.Logf("map size: %d", len(svt.mapToPoolBigNode))
				dx, dy, dz := svt.Dx(), svt.Dy(), svt.Dz()
				x := rand.Intn(dx)
				y := rand.Intn(dy)
				z := rand.Intn(dz)
				b.ResetTimer()
				for b.Loop() {
					_, _ = svt.GetMapToPoolBigNode(x, y, z)
					x = (x + 1) % dx
					y = (y + 2) % dy
					z = (z + 4) % dz
				}
			})
			b.Run(fmt.Sprintf("Cache-%v-%v", s, f), func(b *testing.B) {
				dx, dy, dz := svt.Dx(), svt.Dy(), svt.Dz()
				x := rand.Intn(dx)
				y := rand.Intn(dy)
				z := rand.Intn(dz)
				b.ResetTimer()
				for b.Loop() {
					_, _ = svt.GetWithCache(x, y, z)
					x = (x + 1) % dx
					y = (y + 2) % dy
					z = (z + 4) % dz
				}
				b.StopTimer()
				b.Log(svt.CacheStats())
			})
		}
	}
}

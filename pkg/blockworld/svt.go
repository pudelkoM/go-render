package blockworld

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"slices"
	"strings"
)

const mask_5_bits uint64 = 0b11111 // upper 20 bits store x, y, z coordinates
const mask_3_bits = 0b111          // lowest 3 bits for x, y, z coordinates
const mask_2_bits = 0b11           // upper 2 bits for x, y, z coordinates

type LRUitem[K comparable, V any] struct {
	key K
	val V
	c   uint16
}

type LRU[K comparable, V any] struct {
	size  int
	items []LRUitem[K, V]
}

func NewLRU[K comparable, V any](size int) (*LRU[K, V], error) {
	return &LRU[K, V]{
		size:  size,
		items: make([]LRUitem[K, V], size),
	}, nil
}

func (c *LRU[K, V]) Get(key K) (value V, ok bool) {
	idx := slices.IndexFunc(c.items, func(item LRUitem[K, V]) bool {
		return item.key == key && item.c > 0
	})
	if idx == -1 {
		return value, false // Key not found
	}
	newc := c.items[idx].c + 1 // Increment the access count
	if newc < math.MaxUint16-1 {
		c.items[idx].c = newc // Update the access count
	}

	return c.items[idx].val, true
}

func (c *LRU[K, V]) Add(key K, value V) (evicted bool) {
	slices.SortFunc(c.items, func(a, b LRUitem[K, V]) int {
		if a.c == b.c {
			return 0 // Keep the order if counts are equal
		}
		if a.c > b.c {
			return -1 // a comes before b
		}
		return 1 // a comes after b
	})
	idx := slices.IndexFunc(c.items, func(item LRUitem[K, V]) bool {
		return item.key == key
	})
	if idx > 0 {
		return false
	} else {
		c.items[len(c.items)-1] = LRUitem[K, V]{key: key, val: value, c: 1}
		return true // Evict the least recently used item
	}
}

type SVT struct {
	dx, dy, dz       int
	m                map[uint64]*node0
	mapToBigNode     map[uint64]*bigNode
	mapToPoolBigNode map[uint64]uint32
	bigNodePool      []bigNode

	// l1         []*node0
	l1 []*bigNode
	// cache      *simplelru.LRU[uint64, *node0]
	cache *LRU[uint64, *bigNode]
	// cache        *lfucache.Cache[uint64, *node0] // Cache for nodes
	cache_hits   uint64
	cache_misses uint64
	cache_sets   uint64
}

func mapIdx(x, y, z int) uint64 {
	return (uint64(x)&^mask_5_bits)<<0 |
		(uint64(y)&^mask_5_bits)<<20 |
		(uint64(z)&^mask_5_bits)<<40
}

func reverseMapIdx(idx uint64) (x, y, z int) {
	x = int((idx >> 0) & 0xFFFFF)  // Extract the lower 20 bits for x
	y = int((idx >> 20) & 0xFFFFF) // Extract the next 20 bits for y
	z = int((idx >> 40) & 0xFFFFF) // Extract the upper 20 bits for z
	return x, y, z
}

func l1Index(x, y, z int) uint64 {
	const mask = 0b1111111100000 // Keep upper 8 bits (out of 8+5)
	x &= mask
	y &= mask
	z &= mask
	return uint64(
		((x >> 5) << 0) |
			((y >> 5) << 8) |
			((z >> 5) << 16))
}

func bigNodeIndex(x, y, z int) uint64 {
	const mask = 0b11111111000000 // Mask lower 6 bits
	x &= mask
	y &= mask
	z &= mask
	return uint64(
		((x >> 6) << 0) |
			((y >> 6) << 8) |
			((z >> 6) << 16))
}

func (svt *SVT) String() string {
	if len(svt.m) == 0 {
		return "SVT(empty)"
	}

	sb := &strings.Builder{}
	sb.WriteString("SVT(\n")
	for k, _ := range svt.m {
		x, y, z := reverseMapIdx(k)
		sb.WriteString(fmt.Sprintf("Chunk(%v,%v,%v)[%v]: %s\n", x, y, z, k, svt.m[k].String("	")))
	}
	sb.WriteString(")\n")

	return sb.String()

	// return fmt.Sprintf("SVT(\n%v\n)", svt.root.String("	"))
}

func NewSVT(x, y, z int) *SVT {
	// Size check for L1 array
	if x > math.MaxUint16 || y > math.MaxUint16 || z > math.MaxUint16 {
		panic(fmt.Sprintf("SVT dimensions too large: %d x %d x %d exceeds uint16 limit", x, y, z))
	}

	// l, _ := simplelru.NewLRU[uint64, *node0](64, nil)
	l, _ := NewLRU[uint64, *bigNode](64)
	// l, _ := lfucache.New[uint64, *node0](1024 * 16)
	svt := &SVT{
		m: make(map[uint64]*node0, 1024*64),
		// m:     make(map[uint64]*node0),
		// l1:    make([]*node0, 1<<24),
		mapToBigNode:     make(map[uint64]*bigNode, 1024*64),
		mapToPoolBigNode: make(map[uint64]uint32, 1024*64),
		bigNodePool:      make([]bigNode, 0, 1024*64),
		l1:               make([]*bigNode, 1<<24),
		dx:               x,
		dy:               y,
		dz:               z,
		cache:            l,
	}
	return svt
}

func (svt *SVT) Reconcile() {
	// Reconcile all nodes in the map
	for _, n := range svt.m {
		n.Reconcile()
	}
	// for i := range svt.l1 {
	// 	if svt.l1[i] != nil {
	// 		svt.l1[i].Reconcile()
	// 	}
	// }
}

func (svt *SVT) RandomFill(fillPerc float64) {
	if fillPerc < 0.01 {
		num_set_blocks := float64(svt.dx) * float64(svt.dy) * float64(svt.dz) * fillPerc
		// fmt.Println("Filling SVT with", num_set_blocks, "blocks (fill percentage:", fillPerc, ")")
		for range uint64(num_set_blocks) {
			x := rand.Intn(svt.dx)
			y := rand.Intn(svt.dy)
			z := rand.Intn(svt.dz)
			for {
				_, found := svt.Get(x, y, z)
				if !found {
					break
				}
				// If the block is already set, find a new random position
				x = rand.Intn(svt.dx)
				y = rand.Intn(svt.dy)
				z = rand.Intn(svt.dz)
			}
			b := Block{
				Color: color.NRGBA{R: uint8(rand.Intn(256)), G: uint8(rand.Intn(256)), B: uint8(rand.Intn(256)), A: 255},
				IsSet: true,
			}
			svt.Set(x, y, z, b)
			svt.SetArr(x, y, z, b)
			svt.SetMapToBigNode(x, y, z, b)
			svt.SetMapToPoolBigNode(x, y, z, b)
		}
	} else {
		for z := range svt.dz {
			for y := range svt.dy {
				for x := range svt.dx {
					if rand.Float64() < fillPerc {
						b := Block{
							Color: color.NRGBA{R: uint8(rand.Intn(256)), G: uint8(rand.Intn(256)), B: uint8(rand.Intn(256)), A: 255},
							IsSet: true,
						}
						svt.Set(x, y, z, b)
						svt.SetArr(x, y, z, b)
						svt.SetMapToBigNode(x, y, z, b)
						svt.SetMapToPoolBigNode(x, y, z, b)
					}
				}
			}
		}
	}

	svt.Reconcile()
}

func (svt *SVT) Dx() int {
	return svt.dx
}

func (svt *SVT) Dy() int {
	return svt.dy
}

func (svt *SVT) Dz() int {
	return svt.dy
}

func (svt *SVT) CacheStats() string {
	return fmt.Sprintf("Cache hits: %d, Cache misses: %d, Cache sets: %d", svt.cache_hits, svt.cache_misses, svt.cache_sets)
}

func (svt *SVT) GetWithCache(x, y, z int) (*Block, bool) {
	if x < 0 || x >= svt.dx || y < 0 || y >= svt.dy || z < 0 || z >= svt.dz {
		return nil, false
	}
	// idx := mapIdx(x, y, z)
	idx := bigNodeIndex(x, y, z)
	n, ok := svt.cache.Get(idx)
	if ok {
		// if ok == nil {
		svt.cache_hits++
		// If the node is in the cache, return the relative block
		return n.Get(x, y, z)
	} else {
		svt.cache_misses++
	}

	n, _ = svt.mapToBigNode[idx]
	// svt.cache.Set(idx, n)
	svt.cache.Add(idx, n)
	svt.cache_sets++
	return n.Get(x, y, z)

	// if n, ok := svt.m[idx]; ok {
	// 	// If the node exists in the map, get the relative block
	// 	b, ok := n.GetRelative(x, y, z)
	// 	if ok {
	// 		// svt.cache.Add(idx, n)
	// 		svt.cache.Set(idx, n)
	// 		svt.cache_sets++
	// 	}
	// 	return b, ok
	// } else {
	// 	svt.cache.Set(idx, n)
	// 	svt.cache_sets++
	// 	return nil, false
	// }
}

func (svt *SVT) Get(x, y, z int) (*Block, bool) {
	if x < 0 || x >= svt.dx || y < 0 || y >= svt.dy || z < 0 || z >= svt.dz {
		return nil, false
	}
	idx := mapIdx(x, y, z)
	if n, ok := svt.m[idx]; ok {
		// If the node exists in the map, get the relative block
		return n.GetRelative(x, y, z)
	} else {
		return nil, false
	}
}

func (svt *SVT) GetArrBigNode(x, y, z int) (*Block, bool) {
	if x < 0 || x >= svt.dx || y < 0 || y >= svt.dy || z < 0 || z >= svt.dz {
		return nil, false
	}
	idx := bigNodeIndex(x, y, z)
	return svt.l1[idx].Get(x, y, z)
}

func (svt *SVT) GetMapToBigNode(x, y, z int) (*Block, bool) {
	if x < 0 || x >= svt.dx || y < 0 || y >= svt.dy || z < 0 || z >= svt.dz {
		return nil, false
	}
	idx := bigNodeIndex(x, y, z)
	if n, ok := svt.mapToBigNode[idx]; ok {
		return n.Get(x, y, z)
	} else {
		return nil, false
	}
}

func (svt *SVT) GetMapToPoolBigNode(x, y, z int) (*Block, bool) {
	if x < 0 || x >= svt.dx || y < 0 || y >= svt.dy || z < 0 || z >= svt.dz {
		return nil, false
	}
	idx := bigNodeIndex(x, y, z)
	if n, ok := svt.mapToPoolBigNode[idx]; ok {
		return svt.bigNodePool[n].Get(x, y, z)
	} else {
		return nil, false
	}
}

func (svt *SVT) GetInline(x, y, z int) (*Block, bool) {
	if x < 0 || x >= svt.dx || y < 0 || y >= svt.dy || z < 0 || z >= svt.dz {
		return nil, false
	}
	idx := mapIdx(x, y, z)
	n, ok := svt.m[idx]
	if !ok {
		return nil, false
	}
	if n.IsEmpty() {
		return nil, false
	}
	localX := (x >> 3) & mask_2_bits
	localY := (y >> 3) & mask_2_bits
	localZ := (z >> 3) & mask_2_bits
	index := (localZ << 4) | (localY << 2) | localX // Calculate the index in the leafs array
	l := &n.leafs[index]
	if l.IsEmpty() {
		return nil, false
	}
	x &= mask_3_bits
	y &= mask_3_bits
	z &= mask_3_bits
	index = (z << 6) | (y << 3) | x
	return &l.blocks[index], l.blocks[index].IsSet
}

func (svt *SVT) GetJustRangeCheck(x, y, z int) (*Block, bool) {
	if x < 0 || x >= svt.dx || y < 0 || y >= svt.dy || z < 0 || z >= svt.dz {
		return nil, false
	}
	return nil, true
}

func (svt *SVT) Set(x, y, z int, b Block) {
	if x < 0 || x >= svt.dx || y < 0 || y >= svt.dy || z < 0 || z >= svt.dz {
		return // Out of bounds
	}

	idx := mapIdx(x, y, z)
	if n, ok := svt.m[idx]; ok {
		n.Set(x, y, z, b)
	} else {
		n := &node0{}
		n.Set(x, y, z, b)
		svt.m[idx] = n
	}
}

func (svt *SVT) SetArr(x, y, z int, b Block) {
	if x < 0 || x >= svt.dx || y < 0 || y >= svt.dy || z < 0 || z >= svt.dz {
		return // Out of bounds
	}
	// idx := l1Index(x, y, z)
	idx := bigNodeIndex(x, y, z)
	if svt.l1[idx] == nil {
		// svt.l1[idx] = &node0{}
		svt.l1[idx] = &bigNode{}
	}
	svt.l1[idx].Set(x, y, z, b)
}

func (svt *SVT) SetMapToBigNode(x, y, z int, b Block) {
	if x < 0 || x >= svt.dx || y < 0 || y >= svt.dy || z < 0 || z >= svt.dz {
		return // Out of bounds
	}
	idx := bigNodeIndex(x, y, z)
	if n, ok := svt.mapToBigNode[idx]; ok {
		n.Set(x, y, z, b)
	} else {
		n := &bigNode{}
		n.Set(x, y, z, b)
		svt.mapToBigNode[idx] = n
	}
}

func (svt *SVT) SetMapToPoolBigNode(x, y, z int, b Block) {
	if x < 0 || x >= svt.dx || y < 0 || y >= svt.dy || z < 0 || z >= svt.dz {
		return // Out of bounds
	}
	idx := bigNodeIndex(x, y, z)
	if n, ok := svt.mapToPoolBigNode[idx]; ok {
		svt.bigNodePool[n].Set(x, y, z, b)
	} else {
		svt.bigNodePool = append(svt.bigNodePool, bigNode{})
		n := &svt.bigNodePool[len(svt.bigNodePool)-1]
		n.Set(x, y, z, b)
		svt.mapToPoolBigNode[idx] = uint32(len(svt.bigNodePool) - 1)
	}
}

type node0 struct {
	leafs [64]leaf // 4x4x4 leaf layout (x * y * z)
	flags uint8    // flags: bit 0 indicates if the leaf is empty (0) or not (1)
}

func (n *node0) String(indent string) string {
	if n.IsEmpty() {
		return "node0(empty)"
	}

	sb := &strings.Builder{}
	sb.WriteString(indent)
	sb.WriteString("node0(\n")
	for i := range n.leafs {
		sb.WriteString(indent)
		sb.WriteString(n.leafs[i].String(indent + "	"))
		sb.WriteString("\n")
	}
	sb.WriteString(indent)
	sb.WriteString(")")

	return sb.String()
}

func (n *node0) Dx() int {
	return n.leafs[0].Dx() * 4
}

func (n *node0) Dy() int {
	return n.leafs[0].Dy() * 4
}

func (n *node0) Dz() int {
	return n.leafs[0].Dz() * 4
}

func (n *node0) GetRelative(x, y, z int) (*Block, bool) {
	if n == nil {
		return nil, false // Node is nil, return false
	}
	if n.IsEmpty() {
		return nil, false
	}
	localX := (x >> 3) & mask_2_bits
	localY := (y >> 3) & mask_2_bits
	localZ := (z >> 3) & mask_2_bits

	index := (localZ << 4) | (localY << 2) | localX // Calculate the index in the leafs array

	return n.leafs[index].GetRelative(x, y, z)
}

func (n *node0) GetInline(x, y, z int) (*Block, bool) {
	if n.IsEmpty() {
		return nil, false
	}
	localX := (x >> 3) & mask_2_bits
	localY := (y >> 3) & mask_2_bits
	localZ := (z >> 3) & mask_2_bits

	index := (localZ << 4) | (localY << 2) | localX // Calculate the index in the leafs array

	l := &n.leafs[index]
	if l.IsEmpty() {
		return nil, false
	}
	x &= mask_3_bits
	y &= mask_3_bits
	z &= mask_3_bits
	index = (z << 6) | (y << 3) | x
	return &l.blocks[index], l.blocks[index].IsSet
}

func (n *node0) Set(x, y, z int, b Block) {
	localX := (x >> 3) & mask_2_bits
	localY := (y >> 3) & mask_2_bits
	localZ := (z >> 3) & mask_2_bits

	index := (localZ << 4) | (localY << 2) | localX // Calculate the index in the leafs array

	n.leafs[index].Set(x, y, z, b)
}

func (n *node0) IsEmpty() bool {
	return n.flags&1 == 0
}

func (n *node0) Reconcile() {
	// Check if all leafs are empty
	allEmpty := true
	for i := range n.leafs {
		n.leafs[i].Reconcile() // Reconcile each leaf
		if !n.leafs[i].IsEmpty() {
			allEmpty = false
		}
	}
	if allEmpty {
		n.flags &^= 1 // Clear the empty flag
	} else {
		n.flags |= 1 // Set the empty flag
	}
}

type leaf struct {
	blocks [512]Block // 8x8x8 block layout (x * y * z)
	flags  uint8      // flags: bit 0 indicates if the leaf is empty (0) or not (1)
}

func (l *leaf) String(indent string) string {
	if l.IsEmpty() {
		return indent + "leaf(empty)"
	}

	s := func(b *Block) string {
		if b.IsSet {
			return "#"
		} else {
			return "O"
		}
	}

	b0 := s(&l.blocks[0])
	b1 := s(&l.blocks[1])
	b2 := s(&l.blocks[2])
	b3 := s(&l.blocks[3])
	b4 := s(&l.blocks[4])
	b5 := s(&l.blocks[5])
	b6 := s(&l.blocks[6])
	b7 := s(&l.blocks[7])
	return fmt.Sprintf("%sleaf(%s%s%s%s%s%s%s%s)", indent, b0, b1, b2, b3, b4, b5, b6, b7)
}

func (l *leaf) Dx() int {
	return 8
}

func (l *leaf) Dy() int {
	return 8
}

func (l *leaf) Dz() int {
	return 8
}

// GetRelative retrieves a block relative to the leaf's local coordinates.
func (l *leaf) GetRelative(x, y, z int) (*Block, bool) {
	if l.IsEmpty() {
		return nil, false
	}
	x &= mask_3_bits
	y &= mask_3_bits
	z &= mask_3_bits

	index := (z << 6) | (y << 3) | x

	return &l.blocks[index], l.blocks[index].IsSet
}

func (l *leaf) Set(x, y, z int, b Block) {
	x &= mask_3_bits
	y &= mask_3_bits
	z &= mask_3_bits
	index := (z << 6) | (y << 3) | x
	l.blocks[index] = b
}

func (l *leaf) IsEmpty() bool {
	return l.flags&1 == 0
}

func (l *leaf) Reconcile() {
	// Check if all blocks are empty
	allEmpty := true
	for i := range l.blocks {
		if l.blocks[i].IsSet {
			allEmpty = false
		}
	}
	if allEmpty {
		l.flags &^= 1 // Clear the empty flag
	} else {
		l.flags |= 1 // Set the empty flag
	}
}

type bigNode struct {
	blocks [262144]Block // 64x64x64 block layout (x * y * z)
}

func (b *bigNode) Get(x, y, z int) (*Block, bool) {
	if b == nil {
		return nil, false // Node is nil, return false
	}

	const mask_6_bits = 0b111111
	// if b.IsEmpty() {
	// 	return nil, false
	// }
	x &= mask_6_bits
	y &= mask_6_bits
	z &= mask_6_bits
	index := (z << 12) | (y << 6) | x
	return &b.blocks[index], b.blocks[index].IsSet
}

func (b *bigNode) Set(x, y, z int, bl Block) {
	const mask_6_bits = 0b111111
	x &= mask_6_bits
	y &= mask_6_bits
	z &= mask_6_bits
	index := (z << 12) | (y << 6) | x
	b.blocks[index] = bl
}

type node32 struct {
	blocks [32768]Block // 32x32x32 block layout (x * y * z)
}

func (b *node32) Get(x, y, z int) (*Block, bool) {
	if b == nil {
		return nil, false // Node is nil, return false
	}
	const mask_5_bits = 0b11111
	x &= mask_5_bits
	y &= mask_5_bits
	z &= mask_5_bits
	index := (z << 10) | (y << 5) | x
	return &b.blocks[index], b.blocks[index].IsSet
}

func (b *node32) Set(x, y, z int, bl Block) {
	const mask_5_bits = 0b11111
	x &= mask_5_bits
	y &= mask_5_bits
	z &= mask_5_bits
	index := (z << 10) | (y << 5) | x
	b.blocks[index] = bl
}

type node16 struct {
	blocks [4096]Block // 16x16x16 block layout (x * y * z)
}

func (b *node16) Get(x, y, z int) (*Block, bool) {
	if b == nil {
		return nil, false // Node is nil, return false
	}
	const mask_4_bits = 0b1111
	x &= mask_4_bits
	y &= mask_4_bits
	z &= mask_4_bits
	index := (z << 8) | (y << 4) | x
	return &b.blocks[index], b.blocks[index].IsSet
}

func (b *node16) Set(x, y, z int, bl Block) {
	const mask_4_bits = 0b1111
	x &= mask_4_bits
	y &= mask_4_bits
	z &= mask_4_bits
	index := (z << 8) | (y << 4) | x
	b.blocks[index] = bl
}

// Idea for sparse, but flat node array:
//  - Flat node/block pool
//  - map[xyz]->index into pool

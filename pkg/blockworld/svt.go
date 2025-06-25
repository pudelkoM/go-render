package blockworld

import (
	"fmt"
	"image/color"
	"math/rand"
	"strings"
)

type SVT struct {
	root *node0
}

func (svt *SVT) String() string {
	if svt.root.IsEmpty() {
		return "SVT(empty)"
	}

	return fmt.Sprintf("SVT(\n%v\n)", svt.root.String("	"))
}

func NewSVT() *SVT {
	svt := &SVT{
		root: &node0{},
	}
	return svt
}

func (svt *SVT) RandomFill(fillPerc float64) {
	svt.root = &node0{}
	for i := 0; i < len(svt.root.leafs); i++ {
		svt.root.leafs[i] = node1{}
		if rand.Float64() < fillPerc {
			for j := 0; j < len(svt.root.leafs[i].leafs); j++ {
				svt.root.leafs[i].leafs[j] = node2{}
				if rand.Float64() < fillPerc {
					for k := 0; k < len(svt.root.leafs[i].leafs[j].leafs); k++ {
						svt.root.leafs[i].leafs[j].leafs[k] = leaf{}
						if rand.Float64() < fillPerc {
							for l := 0; l < len(svt.root.leafs[i].leafs[j].leafs[k].blocks); l++ {
								if rand.Float64() < fillPerc {
									svt.root.leafs[i].leafs[j].leafs[k].blocks[l] = Block{
										Color: color.NRGBA{R: uint8(rand.Intn(256)), G: uint8(rand.Intn(256)), B: uint8(rand.Intn(256)), A: 255},
										IsSet: true,
									}
								}
							}
						}
					}
				}
			}
		}
	}

	svt.root.Reconcile()
}

func (svt *SVT) Dx() int {
	return svt.root.Dx()
}

func (svt *SVT) Dy() int {
	return svt.root.Dy()
}

func (svt *SVT) Dz() int {
	return svt.root.Dz()
}

func (svt *SVT) Get(x, y, z int) (Block, bool) {
	return svt.root.GetRelative(x, y, z)
}

type node0 struct {
	leafs [64]node1 // 4x4x4 leaf layout (x * y * z)
	flags uint8     // flags: bit 0 indicates if the leaf is empty (0) or not (1)
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

func (n *node0) GetRelative(x, y, z int) (Block, bool) {
	if n.IsEmpty() {
		return Block{}, false
	}
	if x < 0 || x >= n.Dx() || y < 0 || y >= n.Dy() || z < 0 || z >= n.Dz() {
		return Block{}, false
	}

	// Calculate the leaf index
	nDx := n.leafs[0].Dx()
	nDy := n.leafs[0].Dy()
	nDz := n.leafs[0].Dz()

	xidx := x / nDx
	yidx := y / nDy
	zidx := z / nDz

	return n.leafs[xidx*4+yidx*2+zidx].GetRelative(x%nDx, y%nDy, z%nDz)
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

type node1 struct {
	leafs [64]node2 // 4x4x4 leaf layout (x * y * z)
	flags uint8     // flags: bit 0 indicates if the leaf is empty (0) or not (1)
}

func (n *node1) String(indent string) string {
	if n.IsEmpty() {
		return indent + "node1(empty)"
	}

	sb := &strings.Builder{}
	sb.WriteString(indent)
	sb.WriteString("node1(\n")
	for i := range n.leafs {
		sb.WriteString(indent)
		sb.WriteString(n.leafs[i].String(indent + "	"))
		sb.WriteString("\n")
	}
	sb.WriteString(indent)
	sb.WriteString(")")

	return sb.String()
}

func (n *node1) Dx() int {
	return n.leafs[0].Dx() * 4
}

func (n *node1) Dy() int {
	return n.leafs[0].Dy() * 4
}

func (n *node1) Dz() int {
	return n.leafs[0].Dz() * 4
}

func (n *node1) GetRelative(x, y, z int) (Block, bool) {
	if n.IsEmpty() {
		return Block{}, false
	}
	if x < 0 || x >= n.Dx() || y < 0 || y >= n.Dy() || z < 0 || z >= n.Dz() {
		return Block{}, false
	}

	// Calculate the leaf index
	nDx := n.leafs[0].Dx()
	nDy := n.leafs[0].Dy()
	nDz := n.leafs[0].Dz()

	xidx := x / nDx
	yidx := y / nDy
	zidx := z / nDz

	return n.leafs[xidx*4+yidx*2+zidx].GetRelative(x%nDx, y%nDy, z%nDz)
}

func (n *node1) IsEmpty() bool {
	return n.flags&1 == 0
}

func (n *node1) Reconcile() {
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

type node2 struct {
	leafs [64]leaf // 4x4x4 leaf layout (x * y * z)
	flags uint8    // flags: bit 0 indicates if the leaf is empty (0) or not (1)
}

func (n *node2) String(indent string) string {
	if n.IsEmpty() {
		return indent + "node2(empty)"
	}

	sb := &strings.Builder{}
	sb.WriteString(indent)
	sb.WriteString("node2(\n")
	for i := range n.leafs {
		sb.WriteString(indent)
		sb.WriteString(n.leafs[i].String(indent + "	"))
		sb.WriteString("\n")
	}
	sb.WriteString(indent)
	sb.WriteString(")")

	return sb.String()
}

func (n *node2) Dx() int {
	return n.leafs[0].Dx() * 4
}

func (n *node2) Dy() int {
	return n.leafs[0].Dy() * 4
}

func (n *node2) Dz() int {
	return n.leafs[0].Dz() * 4
}

func (n *node2) GetRelative(x, y, z int) (Block, bool) {
	if n.IsEmpty() {
		return Block{}, false
	}
	if x < 0 || x >= n.Dx() || y < 0 || y >= n.Dy() || z < 0 || z >= n.Dz() {
		return Block{}, false
	}

	// Calculate the leaf index
	nDx := n.leafs[0].Dx()
	nDy := n.leafs[0].Dy()
	nDz := n.leafs[0].Dz()

	xidx := x / nDx
	yidx := y / nDy
	zidx := z / nDz

	return n.leafs[xidx*4+yidx*2+zidx].GetRelative(x%nDx, y%nDy, z%nDz)
}

func (n *node2) IsEmpty() bool {
	return n.flags&1 == 0
}

func (n *node2) Reconcile() {
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
func (l *leaf) GetRelative(x, y, z int) (Block, bool) {
	if l.IsEmpty() {
		return Block{}, false
	}
	if x < 0 || x >= l.Dx() || y < 0 || y >= l.Dy() || z < 0 || z >= l.Dz() {
		return Block{}, false
	}
	return l.blocks[x*4+y*2+z], true
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

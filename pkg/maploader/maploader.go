package maploader

import (
	"encoding/binary"
	"errors"
	"image/color"
	"os"

	"github.com/pudelkoM/go-render/pkg/blockworld"
)

var (
	mapData   [512][512][64]int
	colorData [512][512][64]uint32
)

func LoadMap(path string, world *blockworld.Blockworld) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = loadMap(data, len(data))
	if err != nil {
		return err
	}

	world.SetSize(512, 512, 64)

	for x := 0; x < len(mapData); x++ {
		for y := 0; y < len(mapData[x]); y++ {
			for z := 0; z < len(mapData[x][y]); z++ {
				if mapData[x][y][z] == 1 {
					var col color.Color = color.Black
					c := colorData[x][y][z]
					col = color.NRGBA{
						R: uint8((c >> 24) & 0xFF),
						G: uint8((c >> 16) & 0xFF),
						B: uint8((c >> 8) & 0xFF),
						A: uint8(c & 0xFF),
					}
					world.Set(x, y, z, blockworld.Block{
						Color: col,
					})
				}
			}
		}
	}
	return nil
}

func setGeom(x, y, z, t int) error {
	z = 63 - z
	if z < 0 || z >= 64 {
		return errors.New("z out of bounds")
	}
	mapData[x][y][z] = t
	return nil
}

func setColor(x, y, z int, c uint32) error {
	z = 63 - z
	if z < 0 || z >= 64 {
		return errors.New("z out of bounds")
	}
	colorData[x][y][z] = c
	return nil
}

func loadMap(v []byte, length int) error {
	// base := v
	var x, y, z int

	for y = 0; y < 512; y++ {
		for x = 0; x < 512; x++ {
			for z = 0; z < 64; z++ {
				if err := setGeom(x, y, z, 1); err != nil {
					return err
				}
			}
			z = 0
			for {
				if len(v) < 4 {
					return errors.New("insufficient data")
				}
				number4ByteChunks := int(v[0])
				topColorStart := int(v[1])
				topColorEnd := int(v[2]) // inclusive

				for i := z; i < topColorStart; i++ {
					if err := setGeom(x, y, i, 0); err != nil {
						return err
					}
				}

				color := v[4:]
				for z = topColorStart; z <= topColorEnd; z++ {
					if len(color) < 4 {
						return errors.New("insufficient color data")
					}
					c := binary.LittleEndian.Uint32(color)
					if err := setColor(x, y, z, c); err != nil {
						return err
					}
					color = color[4:]
				}

				lenBottom := topColorEnd - topColorStart + 1

				if number4ByteChunks == 0 {
					v = v[4*(lenBottom+1):]
					break
				}

				lenTop := (number4ByteChunks - 1) - lenBottom

				v = v[number4ByteChunks*4:]

				if len(v) < 4 {
					return errors.New("insufficient data for bottom color end")
				}
				bottomColorEnd := int(v[3])
				bottomColorStart := bottomColorEnd - lenTop

				for z = bottomColorStart; z < bottomColorEnd; z++ {
					if len(color) < 4 {
						return errors.New("insufficient color data")
					}
					c := binary.LittleEndian.Uint32(color)
					if err := setColor(x, y, z, c); err != nil {
						return err
					}
					color = color[4:]
				}
			}
		}
	}

	// if len(v) != length {
	// 	return fmt.Errorf("length mismatch: %d != %d", len(v), length)
	// }

	return nil
}

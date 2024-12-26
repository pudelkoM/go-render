package utils

import (
	"image/color"
)

func CompositeNRGBA(c1 color.NRGBA, c2 color.NRGBA) color.NRGBA {
	return color.NRGBA{
		R: uint8((int(c1.R) * int(c1.A) / 255) + (int(c2.R) * int(c2.A) * (255 - int(c1.A)) / (255 * 255))),
		G: uint8((int(c1.G) * int(c1.A) / 255) + (int(c2.G) * int(c2.A) * (255 - int(c1.A)) / (255 * 255))),
		B: uint8((int(c1.B) * int(c1.A) / 255) + (int(c2.B) * int(c2.A) * (255 - int(c1.A)) / (255 * 255))),
		A: uint8(c1.A + c2.A*(255-c1.A)/255),
	}
}

func ColorDarken(c color.NRGBA, factor float64) color.NRGBA {
	// c.R = uint8(float64(c.R) * factor)
	// c.G = uint8(float64(c.G) * factor)
	// c.B = uint8(float64(c.B) * factor)

	// c.A = uint8(float64(c.A) * factor)
	c.A = uint8(255. * factor)

	// c.R -= factor
	// c.G -= factor
	// c.B -= factor
	// c.A -= factor
	return c
}

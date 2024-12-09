package utils

import "image/color"

func CompositeNRGBA(c1 color.RGBA, c2 color.NRGBA) color.NRGBA {
	return color.NRGBA{
		R: uint8((int(c1.R) * int(c1.A) / 255) + (int(c2.R) * int(c2.A) * (255 - int(c1.A)) / (255 * 255))),
		G: uint8((int(c1.G) * int(c1.A) / 255) + (int(c2.G) * int(c2.A) * (255 - int(c1.A)) / (255 * 255))),
		B: uint8((int(c1.B) * int(c1.A) / 255) + (int(c2.B) * int(c2.A) * (255 - int(c1.A)) / (255 * 255))),
		A: uint8(c1.A + c2.A*(255-c1.A)/255),
	}
}

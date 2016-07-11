package image

import (
	"image"
	"image/draw"

	"github.com/disintegration/imaging"
)

func Resize(img image.Image, newWidth, newHeight int, bestfit bool) image.Image {
	if bestfit {
		img = imaging.Thumbnail(img, newWidth, newHeight, imaging.Lanczos)
	} else {
		var (
			bg            *image.RGBA
			originalRatio float64 = float64(img.Bounds().Dx()) / float64(img.Bounds().Dy())
			newRatio      float64 = float64(newWidth) / float64(newHeight)
			tmpWidth      int     = newWidth
			tmpHeight     int     = newHeight
		)

		if originalRatio > newRatio {
			tmpHeight = 0
		} else {
			tmpWidth = 0
		}

		img = imaging.Resize(img, tmpWidth, tmpHeight, imaging.Lanczos)
		bg = image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.Draw(bg, bg.Bounds(), &image.Uniform{getBgColor()}, image.ZP, draw.Src)

		img = imaging.PasteCenter(bg, img)
	}

	return img
}

package image

import (
	"github.com/rainycape/magick"
	"math"
)

func Resize(img *magick.Image, newWidth, newHeight int, bestfit bool) *magick.Image {
	var (
		err                 error
		oldWidth            int = img.Width()
		oldHeight           int = img.Height()
		tmpWidth, tmpHeight int
		oldRatio            float64 = float64(oldWidth) / float64(oldHeight)
		newRatio            float64 = float64(newWidth) / float64(newHeight)
	)

	if bestfit {
		img, err = img.CropResize(newWidth, newHeight, magick.FQuadratic, magick.CSCenter)

		if err != nil {
			panic(err)
		}
	} else {
		if oldRatio > newRatio {
			// Ширина больше
			tmpWidth = newWidth
			tmpHeight = int(float64(newWidth) / oldRatio)
		} else {
			// Высота больше
			tmpHeight = newHeight
			tmpWidth = int(float64(newHeight) * oldRatio)
		}

		img, err = img.Sample(tmpWidth, tmpHeight)

		if err != nil {
			panic(err)
		}

		borderWidth := float64((newWidth - img.Width()) / 2)
		borderHeight := float64((newHeight - img.Height()) / 2)

		borderWidth = math.Max(0.0, borderWidth)
		borderHeight = math.Max(0.0, borderHeight)

		rect := magick.Rect{
			Width:  uint(borderWidth),
			Height: uint(borderHeight),
			X:      0,
			Y:      0,
		}

		img, err = img.AddBorder(rect, getBgColor())

		if err != nil {
			panic(err)
		}

	}

	return img
}

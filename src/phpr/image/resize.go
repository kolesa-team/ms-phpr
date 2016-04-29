package image

import (
	"github.com/rainycape/magick"
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
		if newRatio > 1 {
			// Ширина больше
			tmpHeight = newHeight
			tmpWidth = int(float64(newHeight) * oldRatio)
		} else {
			// Высота больше
			tmpWidth = newWidth
			tmpHeight = int(float64(newWidth) / oldRatio)
		}

		img, err = img.Sample(tmpWidth, tmpHeight)

		if err != nil {
			panic(err)
		}

		rect := magick.Rect{
			Width:  uint((newWidth - img.Width()) / 2),
			Height: uint((newHeight - img.Height()) / 2),
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

package image

import (
	"math"

	"../logger"

	log "github.com/Sirupsen/logrus"
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
		if img, err = img.CropResize(newWidth, newHeight, magick.FSinc, magick.CSCenter); err != nil {
			logger.Instance().WithFields(log.Fields{
				"error": err,
			}).Info("Resize failed")
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

		if img, err = img.Sample(tmpWidth, tmpHeight); err != nil {
			logger.Instance().WithFields(log.Fields{
				"error": err,
			}).Info("Resample failed")
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

		if img, err = img.AddBorder(rect, getBgColor()); err != nil {
			logger.Instance().WithFields(log.Fields{
				"error": err,
			}).Info("Add border failed")
		}
	}

	return img
}

package image

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"sync"

	"../helper"
	"../logger"
	log "github.com/Sirupsen/logrus"
	"github.com/disintegration/imaging"
	"github.com/endeveit/go-snippets/config"
)

var (
	err                                                                  error
	sizeThreshold, smallSize, bigSize, watermarkSize, watermarkPos       image.Rectangle
	sizeSmallWidth, sizeSmallHeight, sizeBigWidth, sizeBigHeight         int
	colorThreshold                                                       color.Color
	watermarkMargin                                                      int
	watermarkPath                                                        string
	fileBlackBig, fileBlackSmall, fileWhiteBig, fileWhiteSmall, filename string
	once                                                                 sync.Once
	watermarks                                                           map[string]image.Image
)

const (
	CORNER_LEFT_TOP     = 0
	CORNER_RIGHT_TOP    = 1
	CORNER_LEFT_BOTTOM  = 2
	CORNER_RIGHT_BOTTOM = 3
)

func Watermark(image image.Image) image.Image {
	initConfig()

	watermarkSize = pickSize(image)
	watermarkPos = watermarkSize

	switch rand.Intn(4) {
	case CORNER_LEFT_TOP:
		watermarkPos.Max.X = watermarkMargin
		watermarkPos.Max.Y = watermarkMargin
		break
	case CORNER_RIGHT_TOP:
		watermarkPos.Max.X = image.Bounds().Dx() - (int(watermarkSize.Dx()) + watermarkMargin)
		watermarkPos.Max.Y = watermarkMargin
		break
	case CORNER_LEFT_BOTTOM:
		watermarkPos.Max.X = watermarkMargin
		watermarkPos.Max.Y = image.Bounds().Dy() - (int(watermarkSize.Dy()) + watermarkMargin)
		break
	case CORNER_RIGHT_BOTTOM:
		watermarkPos.Max.X = image.Bounds().Dx() - (int(watermarkSize.Dx()) + watermarkMargin)
		watermarkPos.Max.Y = image.Bounds().Dy() - (int(watermarkSize.Dy()) + watermarkMargin)
		break
	}

	color := pickColor(image, watermarkSize)

	switch {
	case color == "b" && watermarkSize.Dx() == bigSize.Dx() && watermarkSize.Dy() == bigSize.Dy():
		filename = fileBlackBig
		break
	case color == "b" && watermarkSize.Dx() == smallSize.Dx() && watermarkSize.Dy() == smallSize.Dy():
		filename = fileBlackSmall
		break
	case color == "w" && watermarkSize.Dx() == bigSize.Dx() && watermarkSize.Dy() == bigSize.Dy():
		filename = fileWhiteBig
		break
	case color == "w" && watermarkSize.Dx() == smallSize.Dx() && watermarkSize.Dy() == smallSize.Dy():
		filename = fileWhiteSmall
		break
	}

	if wm, exists := watermarks[filename]; exists {
		image = imaging.Overlay(image, wm, watermarkPos.Max, 1.0)
	} else {
		logger.Instance().WithFields(log.Fields{
			"filename": filename,
		}).Error("File not found in watermarks set")
	}

	return image
}

func pickColor(img image.Image, rect image.Rectangle) (col string) {
	var (
		sImage *image.NRGBA
	)

	col = "w"

	sImage = imaging.Crop(img, rect)
	sImage = imaging.Resize(image.Image(sImage), 1, 1, imaging.Lanczos)
	pR, pG, pB, _ := sImage.At(0, 0).RGBA()
	tR, tG, tB, _ := colorThreshold.RGBA()

	if pR > tR && pG > tG && pB > tB {
		col = "b"
	}

	return
}

func pickSize(image image.Image) image.Rectangle {
	if image.Bounds().Dx() < sizeThreshold.Dx() || image.Bounds().Dy() < sizeThreshold.Dy() {
		return smallSize
	}

	return bigSize
}

func initConfig() {
	once.Do(func() {
		var (
			watermark image.Image
			err       error
		)
		watermarks = make(map[string]image.Image)

		watermarkPath, err = config.Instance().String("watermark", "path")
		helper.CheckError(err)

		fileWhiteBig, err = config.Instance().String("watermark", "file_white_big")
		helper.CheckError(err)

		fileWhiteSmall, err = config.Instance().String("watermark", "file_white_small")
		helper.CheckError(err)

		fileBlackBig, err = config.Instance().String("watermark", "file_black_big")
		helper.CheckError(err)

		fileBlackSmall, err = config.Instance().String("watermark", "file_black_small")
		helper.CheckError(err)

		watermarkMargin, err = config.Instance().Int("watermark", "margin")
		helper.CheckError(err)

		sizeThreshold = parseConfigSize("watermark", "size_threshold")

		r, g, b := parseConfigColor("watermark", "color_threshold")
		colorThreshold = color.Color(color.RGBA{
			R: r,
			G: g,
			B: b,
			A: uint8(255),
		})

		for _, filename := range []string{fileWhiteBig, fileWhiteSmall, fileBlackBig, fileBlackSmall} {
			watermark, err = imaging.Open(fmt.Sprintf("%s%s", watermarkPath, filename))
			helper.CheckError(err)

			watermarks[filename] = watermark

			if filename == fileBlackBig {
				bigSize = watermark.Bounds()
			}

			if filename == fileBlackSmall {
				smallSize = watermark.Bounds()
			}
		}
	})
}

package image

import (
	"fmt"
	"math/rand"
	"sync"

	"../helper"
	"github.com/endeveit/go-snippets/config"
	"github.com/rainycape/magick"
)

var (
	err                                                                  error
	sizeThreshold, smallSize, bigSize, watermarkSize                     magick.Rect
	sizeSmallWidth, sizeSmallHeight, sizeBigWidth, sizeBigHeight         int
	colorThreshold                                                       magick.Pixel
	watermarkMargin                                                      int
	watermarkPath                                                        string
	fileBlackBig, fileBlackSmall, fileWhiteBig, fileWhiteSmall, filename string
	once                                                                 sync.Once
	watermarks                                                           map[string]*magick.Image
)

const (
	CORNER_LEFT_TOP     = 0
	CORNER_RIGHT_TOP    = 1
	CORNER_LEFT_BOTTOM  = 2
	CORNER_RIGHT_BOTTOM = 3
)

func Watermark(image *magick.Image) *magick.Image {
	initConfig()

	watermarkSize = pickSize(image)

	switch rand.Intn(4) {
	case CORNER_LEFT_TOP:
		watermarkSize.X = watermarkMargin
		watermarkSize.Y = watermarkMargin
		break
	case CORNER_RIGHT_TOP:
		watermarkSize.X = image.Width() - (int(watermarkSize.Width) + watermarkMargin)
		watermarkSize.Y = watermarkMargin
		break
	case CORNER_LEFT_BOTTOM:
		watermarkSize.X = watermarkMargin
		watermarkSize.Y = image.Height() - (int(watermarkSize.Height) + watermarkMargin)
		break
	case CORNER_RIGHT_BOTTOM:
		watermarkSize.X = image.Width() - (int(watermarkSize.Width) + watermarkMargin)
		watermarkSize.Y = image.Height() - (int(watermarkSize.Height) + watermarkMargin)
		break
	}

	color := pickColor(image, watermarkSize)

	switch {
	case color == "b" && watermarkSize.Width == bigSize.Width && watermarkSize.Height == bigSize.Height:
		filename = fileBlackBig
		break
	case color == "b" && watermarkSize.Width == smallSize.Width && watermarkSize.Height == smallSize.Height:
		filename = fileBlackSmall
		break
	case color == "w" && watermarkSize.Width == bigSize.Width && watermarkSize.Height == bigSize.Height:
		filename = fileWhiteBig
		break
	case color == "w" && watermarkSize.Width == smallSize.Width && watermarkSize.Height == smallSize.Height:
		filename = fileWhiteSmall
		break
	}

	if wm, exists := watermarks[filename]; exists {
		image.Composite(magick.CompositeAtop, wm, watermarkSize.X, watermarkSize.Y)
	}

	return image
}

func pickColor(image *magick.Image, rect magick.Rect) (color string) {
	var (
		err    error
		sImage *magick.Image
		pixel  *magick.Pixel
	)

	color = "w"

	sImage, err = image.Clone()
	helper.CheckError(err)

	sImage, err = sImage.Crop(rect)
	helper.CheckError(err)

	sImage, err = sImage.Sample(1, 1)
	helper.CheckError(err)

	pixel, err = sImage.Pixel(0, 0)
	helper.CheckError(err)

	if pixel.Red > colorThreshold.Red && pixel.Green > colorThreshold.Green && pixel.Blue > colorThreshold.Blue {
		color = "b"
	}

	return
}

func pickSize(image *magick.Image) (result magick.Rect) {
	result = bigSize

	if uint(image.Width()) < sizeThreshold.Width || uint(image.Height()) < sizeThreshold.Height {
		result = smallSize
	}

	return result
}

func initConfig() {
	once.Do(func() {
		var (
			w, h      uint64
			watermark *magick.Image
		)

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

		w, h = parseConfigSize("watermark", "size_threshold")
		sizeThreshold = magick.Rect{
			Width:  uint(w),
			Height: uint(h),
		}
		w, h = parseConfigSize("watermark", "size_big")
		bigSize = magick.Rect{
			Width:  uint(w),
			Height: uint(h),
		}
		w, h = parseConfigSize("watermark", "size_small")
		smallSize = magick.Rect{
			Width:  uint(w),
			Height: uint(h),
		}

		r, g, b := parseConfigColor("watermark", "color_threshold")
		colorThreshold = magick.Pixel{
			Red:     r,
			Green:   g,
			Blue:    b,
			Opacity: uint8(255),
		}

		watermarks = make(map[string]*magick.Image)
		for _, filename := range []string{fileWhiteBig, fileWhiteSmall, fileBlackBig, fileBlackSmall} {
			watermark, err = magick.DecodeFile(fmt.Sprintf("%s%s", watermarkPath, filename))
			helper.CheckError(err)

			watermarks[filename] = watermark
		}
	})
}

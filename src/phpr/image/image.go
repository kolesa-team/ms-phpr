package image

import (
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"strconv"
	"strings"

	"../helper"
	"github.com/disintegration/imaging"
	"github.com/endeveit/go-snippets/cli"
	"github.com/endeveit/go-snippets/config"
)

func FromReader(r io.Reader) (image.Image, error) {
	return imaging.Decode(r)
}

func ToWriter(image image.Image, w io.Writer) error {
	return jpeg.Encode(w, image, &jpeg.Options{Quality: 75})
	//	return imaging.Encode(w, image, imaging.JPEG)
}

func getBgColor() color.RGBA {
	red, green, blue := parseConfigColor("image", "background")

	return color.RGBA{
		R: red,
		G: green,
		B: blue,
		A: 255,
	}
}

func parseConfigColor(section, key string) (red, green, blue uint8) {
	var (
		colorStr string
		err      error
		color    []string
	)
	colorStr, err = config.Instance().String(section, key)
	cli.CheckError(err)

	color = strings.Split(colorStr, ",")
	if len(color) != 3 {
		panic("Invalid color string")
	}

	r, _ := strconv.ParseUint(color[0], 10, 8)
	g, _ := strconv.ParseUint(color[1], 10, 8)
	b, _ := strconv.ParseUint(color[2], 10, 8)

	red = uint8(r)
	green = uint8(g)
	blue = uint8(b)

	return
}

func parseConfigSize(section, key string) image.Rectangle {
	var (
		width, height uint64
		size          []string
		value         string
		err           error
	)

	value, err = config.Instance().String(section, key)
	helper.CheckError(err)

	size = strings.Split(value, "x")
	if len(size) != 2 {
		panic("Invalid size string")
	}

	width, err = strconv.ParseUint(size[0], 10, 64)
	helper.CheckError(err)

	height, err = strconv.ParseUint(size[1], 10, 64)
	helper.CheckError(err)

	return image.Rect(0, 0, int(width), int(height))
}

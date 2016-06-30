package image

import (
	"io"
	"strconv"
	"strings"

	"../helper"
	"github.com/endeveit/go-snippets/cli"
	"github.com/endeveit/go-snippets/config"
	"github.com/rainycape/magick"
)

func FromReader(r io.Reader) (*magick.Image, error) {
	return magick.Decode(r)
}

func ToWriter(image *magick.Image, w io.Writer) error {
	info := magick.NewInfo()
	info.SetFormat(image.Format())

	return image.Encode(w, info)
}

func getBgColor() *magick.Pixel {
	red, green, blue := parseConfigColor("image", "background")

	pixel := magick.Pixel{
		Red:     red,
		Green:   green,
		Blue:    blue,
		Opacity: uint8(255),
	}

	return &pixel
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

func parseConfigSize(section, key string) (width, height uint64) {
	var (
		size  []string
		value string
		err   error
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

	return
}

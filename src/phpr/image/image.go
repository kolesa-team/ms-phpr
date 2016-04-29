package image

import (
	"github.com/endeveit/go-snippets/cli"
	"github.com/endeveit/go-snippets/config"
	"github.com/rainycape/magick"
	"io"
	"strconv"
	"strings"
)

func FromReader(r io.Reader) *magick.Image {
	image, err := magick.Decode(r)

	if err != nil {
		panic(err)
	}

	return image
}

func ToWriter(image *magick.Image, w io.Writer) error {
	info := magick.NewInfo()
	info.SetFormat(image.Format())

	return image.Encode(w, info)
}

func getBgColor() *magick.Pixel {
	var (
		colorStr         string
		err              error
		color            []string
		red, green, blue int
	)
	colorStr, err = config.Instance().String("image", "background")
	cli.CheckError(err)

	color = strings.Split(colorStr, ",")

	red, _ = strconv.Atoi(color[0])
	green, _ = strconv.Atoi(color[1])
	blue, _ = strconv.Atoi(color[2])

	pixel := magick.Pixel{
		Red:     uint8(red),
		Green:   uint8(green),
		Blue:    uint8(blue),
		Opacity: uint8(255),
	}

	return &pixel
}

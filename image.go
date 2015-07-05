package goisgod

import (
	"bytes"
	"image"

	// require
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
)

type GigImage struct {
	key   string
	image *image.Image
}

func (gimg *GigImage) toByte() (bs []byte, err error) {

	buf := new(bytes.Buffer)
	if err = jpeg.Encode(buf, *gimg.image, nil); err != nil {
		return
	}

	bs = buf.Bytes()

	return
}

func (gimg *GigImage) fromByte(bs []byte) (err error) {

	var img image.Image

	buf := new(bytes.Buffer)
	if _, err = buf.Write(bs); err != nil {
		return
	}

	if img, err = jpeg.Decode(buf); err != nil {
		return
	}

	gimg.image = &img

	return
}

package goisgod

import (
	"bytes"
	"image"

	// require
	_ "image/gif"
	_ "image/png"
	"image/jpeg"
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

	buf := new(bytes.Buffer)
	if _, err = buf.Read(bs); err != nil {
		return
	}

	if *gimg.image, err = jpeg.Decode(buf); err != nil {
		return
	}

	return
}

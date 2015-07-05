package goisgod

import (
	"image"
	"os"

	"github.com/disintegration/imaging"
	"log"
)

type GopherImageDrawer struct {
	gopher    *image.Image
	dao       *GigDao
	stoppedCh chan struct{}
}

// NewGopherImageDrawer is used to draw gopher via image input channel
func NewGopherImageDrawer(dao *GigDao, stoppingCh <-chan struct{}, rawImageInCh <-chan *GigImage) (gid *GopherImageDrawer, err error) {

	gid = new(GopherImageDrawer)

	if gid.gopher, err = gid.getGopherImage(); err != nil {
		return
	}

	gid.stoppedCh = make(chan struct{}, 1)
	gid.dao = dao

	go func() {

		for {
			select {

			case img := <-rawImageInCh:
				log.Printf("Image incoming")
				if gimg, err := gid.drawGopher(img); err != nil {
					gid.stoppedCh <- struct{}{}
					break
				} else {
					gid.dao.storeImage(gimg, gimg.key)
				}

			case <-stoppingCh:
				gid.stoppedCh <- struct{}{}
				break

			}
		}
	}()

	return
}

func (gid *GopherImageDrawer) getGopherImage() (img *image.Image, err error) {

	img = new(image.Image)

	var f *os.File
	if f, err = os.Open("gopher-normal.png"); err != nil {
		return
	}
	defer f.Close()

	if *img, _, err = image.Decode(f); err != nil {
		return
	}

	return
}

func (gid *GopherImageDrawer) drawGopher(_dst *GigImage) (gimg *GigImage, err error) {

	dst := imaging.Resize(*_dst.image, 300, 0, imaging.Lanczos)

	var img image.Image = imaging.Overlay(dst, *gid.gopher, image.Pt(0, 0), 0.5)

	return &GigImage{image: &img, key: _dst.key}, nil
}

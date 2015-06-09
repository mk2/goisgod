package goisgod

import (
	"image"
	"image/png"
	"log"
	"os"
	"testing"
	"time"

	"github.com/boltdb/bolt"
)

func TestNewSearchImageChan(t *testing.T) {

	t.SkipNow()

	var (
		stopch   = make(chan struct{})
		db       *bolt.DB
		gigimgch <-chan *image.Image
	)

	db, _ = bolt.Open("test.boltdb", 0666, nil)
	gigimgch = NewSearchImageChan(db, stopch)

	time.Sleep(1 * time.Second)

	stopch <- struct{}{}
	<-gigimgch
}

func TestCseSearch(t *testing.T) {

	t.SkipNow()

	var img image.Image

	cseSearch(&img)
}

func TestCseSearchItemToGIGImage(t *testing.T) {

	var res CseSearchItem
	res.Image.ThumbnailLink = "http://ks.c.yimg.jp/res/chie-ans-329/329/831/301/i320"

	gigimg := cseSearchItemToGIGImage(&res)

	if gigimg == nil {
		t.Errorf("Gigimage is nil")
	}
}

func TestDrawGopher(t *testing.T) {

	var res CseSearchItem
	res.Image.ThumbnailLink = "http://ks.c.yimg.jp/res/chie-ans-329/329/831/301/i320"

	gigimg := cseSearchItemToGIGImage(&res)

	goimg := getGopherImage()

	drawGopher(gigimg, goimg)

	f, e := os.Create("go.png")
	defer f.Close()

	log.Printf("Error: ", e)

	png.Encode(f, *gigimg)
}

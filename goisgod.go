package goisgod

import (
	"log"
	"time"
)

type goisgod struct {
	dao  *GigDao
	gid  *GopherImageDrawer
	imgp *ImagePoster
	imgs *ImageSearcher
}

// StartGoIsGod is used to start goisgod bot
func StartGoIsGod() int {

	gig := new(goisgod)

	return gig.loop()
}

func (gig *goisgod) loop() (ret int) {

	ret = 0

	var (
		dao *GigDao
		err error
	)

	if dao, err = NewGigDao("goisgod.boltdb"); err != nil {
		ret = 1
		return
	}

	gig.dao = dao

	imgpStoppingCh := make(chan struct{})
	imgsStoppingCh := make(chan struct{})
	gidStoppingCh := make(chan struct{})
	invokeCh := make(chan struct{})

	gig.imgp = NewImagePoster(dao, imgpStoppingCh, invokeCh)
	gig.imgs = NewImageSearcher(dao, imgsStoppingCh, invokeCh)
	if gig.gid, err = NewGopherImageDrawer(dao, gidStoppingCh, gig.imgs.imageOutCh); err != nil {
		ret = 1
		return
	}

	for {
		c := time.Tick(10 * time.Second)

		for now := range c {

			log.Printf("Now: %s", now)

			// invoke image search routine, image posting routine
			invokeCh <- struct{}{}
		}
	}

	return
}

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
	imgpInvokeCh := make(chan struct{})
	imgsInvokeCh := make(chan struct{})

	gig.imgp = NewImagePoster(dao, imgpStoppingCh, imgpInvokeCh)
	gig.imgs = NewImageSearcher(dao, imgsStoppingCh, imgsInvokeCh)
	if gig.gid, err = NewGopherImageDrawer(dao, gidStoppingCh, gig.imgs.imageOutCh); err != nil {
		ret = 1
		return
	}

	go func() {

		for {
			c := time.Tick(30 * time.Second)

			for now := range c {

				log.Printf("Invoke ImageSearcher: %s", now)

				// invoke image search routine, image posting routine
				imgsInvokeCh <- struct{}{}
			}
		}
	}()

	go func() {
		for {
			c := time.Tick(10 * time.Second)

			for now := range c {

				log.Printf("Invoke ImagePoster: %s", now)

				// invoke image search routine, image posting routine
				imgpInvokeCh <- struct{}{}
			}
		}
	}()

	for {
		select {

		case <-gig.imgp.stoppedCh:
			imgsStoppingCh <- struct{}{}
			gidStoppingCh <- struct{}{}
			break

		case <-gig.imgs.stoppedCh:
			imgpStoppingCh <- struct{}{}
			gidStoppingCh <- struct{}{}
			break

		case <-gig.gid.stoppedCh:
			imgpStoppingCh <- struct{}{}
			gidStoppingCh <- struct{}{}
			break

		}
	}

	return
}

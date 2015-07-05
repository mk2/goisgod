package goisgod

import (
	"bytes"
	"encoding/base64"
	"image/jpeg"
	"log"
	"net/url"

	"github.com/ChimeraCoder/anaconda"
	"github.com/joeshaw/envdecode"
)

// ImagePoster
type ImagePoster struct {
	dao       *GigDao
	api       *anaconda.TwitterApi
	stoppedCh chan struct{}
}

// NewImagePoster is used to post image to twitter
func NewImagePoster(dao *GigDao, stoppingCh <-chan struct{}, invokeCh <-chan struct{}) (imgp *ImagePoster) {

	imgp = new(ImagePoster)

	imgp.dao = dao
	imgp.stoppedCh = make(chan struct{}, 1)

	var env struct {
		ConsumerKey       string `env:"TWITTER_CONSUMER_KEY"`
		ConsumerSecret    string `env:"TWITTER_CONSUMER_SECRET"`
		AccessToken       string `env:"TWITTER_ACCESS_TOKEN"`
		AccessTokenSecret string `env:"TWITTER_ACCESS_TOKEN_SECRET"`
	}

	if err := envdecode.Decode(&env); err != nil {
		log.Printf("Error: %v", err)
	}

	anaconda.SetConsumerKey(env.ConsumerKey)
	anaconda.SetConsumerSecret(env.ConsumerSecret)
	imgp.api = anaconda.NewTwitterApi(env.AccessToken, env.AccessTokenSecret)

	go func() {

		for {
			select {

			case <-invokeCh:
				if err := imgp.post(); err != nil {
					imgp.stoppedCh <- struct{}{}
					break
				}

			case <-stoppingCh:
				imgp.stoppedCh <- struct{}{}
				break

			}
		}

	}()

	return
}

func (imgp *ImagePoster) post() (err error) {

	slack.PostToSlack("Image incoming!!")

	var gimg *GigImage
	if gimg, err = imgp.dao.retrieveRandomImage(); err != nil {
		log.Printf("Error: %v", err)
		return
	}

	imgBuf := new(bytes.Buffer)
	if err := jpeg.Encode(imgBuf, *gimg.image, nil); err != nil {
		log.Fatalf("Jpeg encode failed")
	}

	b64s := base64.StdEncoding.EncodeToString(imgBuf.Bytes())

	media, _ := imgp.api.UploadMedia(b64s)
	v := url.Values{}
	v.Add("media_ids", media.MediaIDString)
	imgp.api.PostTweet("go is god", v)

	slack.PostToSlack("Tweet Posted!!")

	return
}

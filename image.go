package goisgod

import (
	"bytes"
	"encoding/json"
	"image"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	// require
	_ "image/gif"
	"image/jpeg"
	_ "image/png"

	"encoding/base64"

	"errors"
	"fmt"

	"strconv"

	"github.com/ChimeraCoder/anaconda"
	"github.com/boltdb/bolt"
	"github.com/disintegration/imaging"
	"github.com/joeshaw/envdecode"
)

const (
	// CseSearchStartIndexBucket store search start index
	CseSearchStartIndexBucket = "CseSearchStartIndexBucket"

	// CseUsedImageLinkBucket already used iamge index stored
	CseUsedImageLinkBucket = "CseUsedImageLinkBucket"
)

const (

	// CseEndpoint Custom search engine endpoint
	CseEndpoint = "https://www.googleapis.com/customsearch/v1"
)

/*
CseSearchImage Image contained by CseSearchItem
*/
type CseSearchImage struct {
	ContextLink     string `json:"contextLink"`
	Height          int    `json:"height"`
	Width           int    `json:"width"`
	ByteSize        int    `json:"byteSize"`
	ThumbnailLink   string `json:"thumbnailLink"`
	ThumbnailHeight int    `json:"thumbnailHeight"`
	ThumbnailWidth  int    `json:"thumbnailWidth"`
}

/*
CseSearchItem Item contained by CseSearchResult
*/
type CseSearchItem struct {
	Title     string         `json:"title"`
	HTMLTitle string         `json:"htmlTitle"`
	Link      string         `json:"link"`
	Mime      string         `json:"mime"`
	Image     CseSearchImage `json:"image"`
}

/*
CseSearchResult Custom search result
*/
type CseSearchResult struct {
	ResultError map[string]interface{} `json:"error,omitempty"`
	Queries     map[string]interface{} `json:"queries"`
	Items       []CseSearchItem        `json:"items"`
}

/*
StartBot we will start bot!
*/
func StartBot(imgch <-chan *image.Image) {

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
	api := anaconda.NewTwitterApi(env.AccessToken, env.AccessTokenSecret)

	for {
		img := <-imgch
		slack.PostToSlack("Image incoming!!")
		goimg := getGopherImage()
		drawd, _ := drawGopher(img, goimg)

		imgBuf := new(bytes.Buffer)
		if err := jpeg.Encode(imgBuf, drawd, nil); err != nil {
			log.Fatalf("Jpeg encode failed")
			continue
		}

		b64s := base64.StdEncoding.EncodeToString(imgBuf.Bytes())

		media, _ := api.UploadMedia(b64s)
		v := url.Values{}
		v.Add("media_ids", media.MediaIDString)
		api.PostTweet("go is god", v)
		slack.PostToSlack("Tweet Posted!!")
	}
}

/*
NewSearchImageChan Get new image channel
*/
func NewSearchImageChan(db *bolt.DB) <-chan *image.Image {

	db.Update(createSearchImageBuckets)

	ch := make(chan *image.Image)

	go func() {

		var img *image.Image = new(image.Image)

		c := time.Tick(10 * time.Second)

		for now := range c {

			log.Printf("Now: %s", now)

			db.Update(cseSearch(&img))

			if img != nil {
				ch <- img
			}
		}
	}()

	return ch
}

func cseSearch(img **image.Image) func(*bolt.Tx) error {

	// https://www.googleapis.com/customsearch/v1?
	// q=go+is+god&cx=007941957930256544000%3Axrdml4vkxxu&searchType=image&key={YOUR_API_KEY}

	var env struct {
		Cx  string `env:"CSE_CX,required"`
		Key string `env:"CSE_KEY,required"`
	}

	if err := envdecode.Decode(&env); err != nil {
		log.Printf("oh my god")
	}

	return func(tx *bolt.Tx) error {

		var (
			cseItem CseSearchItem
			start   int = 1
			num     int = 10
		)

		for {

			q := url.Values{}
			q.Set("q", "go is god")
			q.Set("cx", env.Cx)
			q.Set("searchType", "image")
			q.Set("key", env.Key)
			q.Set("start", strconv.Itoa(start))

			reqURL := CseEndpoint + "?" + q.Encode()
			log.Printf("reqURL: %s", reqURL)

			resp, _ := http.Get(reqURL)
			bs, _ := ioutil.ReadAll(resp.Body)

			var result CseSearchResult
			json.Unmarshal(bs, &result)

			if result.ResultError != nil {
				log.Printf("An Error occurred: %+v", result)

				return errors.New(fmt.Sprintf("CseError: %+v", result.ResultError))
			}

			log.Printf("CseResult: %+v", result)

			b := tx.Bucket([]byte(CseUsedImageLinkBucket))
			c := b.Cursor()


			for _, i := range result.Items {

				found := false

				cseItem = i

				log.Printf("Item: %v", cseItem)

				prefix := []byte(i.Image.ContextLink)

				for k, _ := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, _ = c.Next() {
					found = true
				}

				if !found {
					*img = cseSearchItemToGIGImage(&cseItem)
					b.Put([]byte(cseItem.Image.ContextLink), []byte(""))
					log.Println("Get Image from cse search item!")
					return nil
				}
			}

			log.Println("Already used")

			start += num
		}

		return nil
	}
}

func drawGopher(_dst *image.Image, _gopher *image.Image) (image.Image, error) {

	dst := imaging.Resize(*_dst, 300, 0, imaging.Lanczos)

	ret := imaging.Overlay(dst, *_gopher, image.Pt(0, 0), 0.5)

	return ret, nil
}

func getGopherImage() *image.Image {

	f, _ := os.Open("gopher-normal.png")
	defer f.Close()
	img, _, _ := image.Decode(f)

	return &img
}

func cseSearchItemToGIGImage(res *CseSearchItem) *image.Image {

	resp, _ := http.Get(res.Image.ThumbnailLink)
	image, _, _ := image.Decode(resp.Body)

	if image == nil {
		log.Printf("image is nil")
	}

	return &image
}

func createSearchImageBuckets(tx *bolt.Tx) error {

	var e error

	_, e = tx.CreateBucketIfNotExists([]byte(CseSearchStartIndexBucket))
	_, e = tx.CreateBucketIfNotExists([]byte(CseUsedImageLinkBucket))

	return e
}

package goisgod

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/joeshaw/envdecode"
)

const (

	// CseEndpoint Custom search engine endpoint
	CseEndpoint = "https://www.googleapis.com/customsearch/v1"
)

// SearchResult
type SearchResult struct {
	imageLink string
	context   string
}

// CseSearchImage Image contained by CseSearchItem
type CseSearchImage struct {
	ContextLink     string `json:"contextLink"`
	Height          int    `json:"height"`
	Width           int    `json:"width"`
	ByteSize        int    `json:"byteSize"`
	ThumbnailLink   string `json:"thumbnailLink"`
	ThumbnailHeight int    `json:"thumbnailHeight"`
	ThumbnailWidth  int    `json:"thumbnailWidth"`
}

// CseSearchItem Item contained by CseSearchResult
type CseSearchItem struct {
	Title     string         `json:"title"`
	HTMLTitle string         `json:"htmlTitle"`
	Link      string         `json:"link"`
	Mime      string         `json:"mime"`
	Image     CseSearchImage `json:"image"`
}

// CseSearchResult Custom search result
type CseSearchResult struct {
	ResultError map[string]interface{} `json:"error,omitempty"`
	Queries     map[string]interface{} `json:"queries"`
	Items       []CseSearchItem        `json:"items"`
}

// ImageSearcher searching image
type ImageSearcher struct {
	dao        *GigDao
	imageOutCh chan *GigImage
	stoppedCh  chan struct{}
}

// NewImageSearcher is used to search image from some search engine
func NewImageSearcher(dao *GigDao, stoppingCh <-chan struct{}, invokeCh <-chan struct{}) (imgs *ImageSearcher) {

	imgs = new(ImageSearcher)

	imgs.dao = dao
	imgs.stoppedCh = make(chan struct{}, 1)
	imgs.imageOutCh = make(chan *GigImage, 1)

	go func() {

		for {
			select {

			case <-invokeCh:
				if gimg, err := imgs.search(); err != nil {
					imgs.stoppedCh <- struct{}{}
					break
				} else {
					log.Printf("Save Image")
					imgs.imageOutCh <- gimg
				}

			case <-stoppingCh:
				imgs.stoppedCh <- struct{}{}
				break

			}
		}

	}()

	return
}

func (imgs *ImageSearcher) search() (*GigImage, error) {

	return imgs.cseSearch()
}

func (imgs *ImageSearcher) cseSearch() (gimg *GigImage, err error) {

	gimg = new(GigImage)

	var (
		cseItem CseSearchItem
		start   int = 1
	)

	var env struct {
		Cx  string `env:"CSE_CX,required"`
		Key string `env:"CSE_KEY,required"`
	}

	if err := envdecode.Decode(&env); err != nil {
		log.Fatalf("google custom engine environment required")
	}


GET_IMAGE:
	for {

		q := url.Values{}
		q.Set("q", "go is god")
		q.Set("cx", env.Cx)
		q.Set("searchType", "image")
		q.Set("key", env.Key)
		q.Set("start", strconv.Itoa(start))

		var reqURL string

		reqURL = CseEndpoint + "?" + q.Encode()
		log.Printf("reqURL: %s", reqURL)

		var resp *http.Response
		if resp, err = http.Get(reqURL); err != nil {
			return
		}

		var bs []byte
		if bs, err = ioutil.ReadAll(resp.Body); err != nil {
			return
		}

		var result CseSearchResult
		if err = json.Unmarshal(bs, &result); err != nil {
			return
		}

		if result.ResultError != nil {
			log.Printf("An Error occurred: %+v", result)
			err = errors.New(fmt.Sprintf("CseError: %+v", result.ResultError))
			return
		}

		// log.Printf("CseResult: %+v", result)

		for _, cseItem = range result.Items {

			// log.Printf("Item: %v", cseItem)

			key := cseItem.Image.ContextLink

			log.Printf("key: %s", key)

			if gimg, err = imgs.dao.retrieveImage(key); err == nil {
				log.Printf("*** already searched image")
				continue
			}
			if gimg.image, err = cseSearchItemToGigImage(&cseItem); err == nil {
				log.Printf("Got new image!")
				gimg.key = key
				break GET_IMAGE
			}
		}

		start += 10
	}

	return
}

func cseSearchItemToGigImage(res *CseSearchItem) (img *image.Image, err error) {

	img = new(image.Image)

	var resp *http.Response
	if resp, err = http.Get(res.Image.ThumbnailLink); err != nil {
		return
	}

	if *img, _, err = image.Decode(resp.Body); err != nil {
		return
	}

	if img == nil {
		return nil, errors.New("image is nil")
	}

	return
}

package goisgod

import (
	"errors"
	"strconv"

	"github.com/boltdb/bolt"
	"log"
	"math/rand"
	"time"
)

// BucketType is used to specify bucket of boltDB
// available convert to []byte via bucketTypeToByte
type BucketType int

const (

	// ImageBucket is used to be stored images
	ImageBucket BucketType = iota
)

// GigDao Database access object in goisgod
type GigDao struct {
	db   *bolt.DB
	path string
}

func bucketTypeToByte(bt BucketType) []byte {
	return []byte("goisgod-" + strconv.Itoa(int(bt)))
}

// NewGigDao return new GigDao
func NewGigDao(path string) (dao *GigDao, err error) {

	var db *bolt.DB
	if db, err = bolt.Open(path, 0666, nil); err != nil {
		return
	}

	dao = &GigDao{
		db:   db,
		path: path,
	}

	return
}

// Close is used to close goisgod database
func (dao *GigDao) Close() error {
	return dao.db.Close()
}

func (dao *GigDao) storeImage(img *GigImage, key string) error {

	log.Printf("will be stored image key: %s", key)

	return dao.db.Update(func(tx *bolt.Tx) (err error) {

		var b *bolt.Bucket
		if b, err = tx.CreateBucketIfNotExists(bucketTypeToByte(ImageBucket)); err != nil {
			return
		}

		var bs []byte
		if bs, err = img.toByte(); err != nil {
			return
		}

		if err = b.Put([]byte(key), bs); err != nil {
			log.Printf("Error: %v", err)
			return
		}

		return nil
	})
}

func (dao *GigDao) retrieveImage(key string) (*GigImage, error) {

	img := new(GigImage)
	err := dao.db.View(func(tx *bolt.Tx) (err error) {

		var b *bolt.Bucket
		if b = tx.Bucket(bucketTypeToByte(ImageBucket)); b == nil {
			return errors.New("bucket not found")
		}

		var bs []byte
		if bs = b.Get([]byte(key)); bs == nil {
			log.Printf("image %s not found", key)
			err = errors.New("not found image of: " + key)
			return
		}


		if err = img.fromByte(bs); err != nil {
			return
		}

		img.key = key

		log.Printf("image %s found", key)

		return nil
	})

	return img, err
}

func randomTrue() bool {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(3)%2 == 0
}

func (dao *GigDao) retrieveRandomImage() (*GigImage, error) {

	img := new(GigImage)
	err := dao.db.Update(func(tx *bolt.Tx) (err error) {

		var b *bolt.Bucket
		if b, err = tx.CreateBucketIfNotExists(bucketTypeToByte(ImageBucket)); err != nil {
			return
		}

		var (
			k  []byte
			bs []byte
		)
		c := b.Cursor()

	RANDOM_IMAGE:
		for {
			for k, bs = c.First(); k != nil; c.Next() {
				if randomTrue() && bs != nil {
					break RANDOM_IMAGE
				}
			}
		}

		log.Printf("key: %v", string(k))

		if err = img.fromByte(bs); err != nil {
			return
		}

		img.key = string(k)

		return nil
	})

	return img, err
}

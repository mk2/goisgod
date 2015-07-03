package goisgod

import (
	"errors"
	"strconv"

	"github.com/boltdb/bolt"
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
			return
		}

		return nil
	})
}

func (dao *GigDao) retreiveImage(key string) (*GigImage, error) {

	img := new(GigImage)
	err := dao.db.Update(func(tx *bolt.Tx) (err error) {

		var b *bolt.Bucket
		if b, err = tx.CreateBucketIfNotExists(bucketTypeToByte(ImageBucket)); err != nil {
			return
		}

		var bs []byte
		if bs = b.Get([]byte(key)); bs == nil {
			err = errors.New("not found image of: " + key)
			return
		}

		if err = img.fromByte(bs); err != nil {
			return
		}

		img.key = key

		return nil
	})

	return img, err
}

package main

import (
	"github.com/boltdb/bolt"
	"github.com/mk2/goisgod"
)

func main() {

	db, _ := bolt.Open("goisgoid.boltdb", 0666, nil)
	defer db.Close()

	go func() {

		ch := goisgod.NewSearchImageChan(db)
		goisgod.StartBot(ch)

	}()
}

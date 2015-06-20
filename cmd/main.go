package main

import (
	"log"

	"github.com/boltdb/bolt"
	"github.com/mk2/goisgod"
)

func main() {

	db, _ := bolt.Open("goisgoid.boltdb", 0666, nil)
	defer db.Close()

	log.Println("Go is God start")
	ch := goisgod.NewSearchImageChan(db)
	goisgod.StartBot(ch)
	log.Println("Go is God End")
}

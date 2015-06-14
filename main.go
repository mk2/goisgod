package main

import (
	"github.com/boltdb/bolt"
	"log"
)

func main() {

	db, _ := bolt.Open("goisgoid.boltdb", 0666, nil)
	defer db.Close()

	log.Println("Go is God start")
	ch := NewSearchImageChan(db)
	StartBot(ch)
	log.Println("Go is God End")
}

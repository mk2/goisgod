package main

import "github.com/boltdb/bolt"

func main() {

	db, _ := bolt.Open("goisgoid.boltdb", 0666, nil)
	defer db.Close()

	go func() {

	}()
}

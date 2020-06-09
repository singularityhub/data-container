package main

import (
	"log"
	"fmt"

	"github.com/tidwall/buntdb"
)

func main() {

	// Open an in-memory database
        db, err := buntdb.Open(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("mykey", "myvalue", nil)
		return err
	})

	db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get("mykey")
		if err != nil{
			return err
		}
		fmt.Printf("value is %s\n", val)
		return nil
	})

}

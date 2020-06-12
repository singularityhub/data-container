package main

import (
	"log"
	"fmt"

	"github.com/vsoch/containerdb"
)

func main() {

	// Open an in-memory database
        db, err := containerdb.Open(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Update(func(tx *containerdb.Tx) error {
		_, _, err := tx.Set("mykey", "myvalue", nil)
		return err
	})

	db.View(func(tx *containerdb.Tx) error {
		val, err := tx.Get("mykey")
		if err != nil{
			return err
		}
		fmt.Printf("value is %s\n", val)
		return nil
	})

}

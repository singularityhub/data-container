package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/vsoch/containerdb"
	"github.com/vsoch/containerdb/gjson"
)

// initdb initializes the database values and indices
func initdb(db *containerdb.DB) error {
	db.CreateIndex("size", "*", containerdb.IndexJSON("size"))
	db.CreateIndex("sha256", "*", containerdb.IndexJSON("sha256"))
	return db.Update(func(tx *containerdb.Tx) error {
		_, _, err := tx.Set(`/data/avocado.txt`, `{"size": 9, "sha256": "327bf8231c9572ecdfdc53473319699e7b8e6a98adf0f383ff6be5b46094aba4"}`, nil)
		_, _, err = tx.Set(`/data/tomato.txt`, `{"size": 8, "sha256": "3b7721618a86990a3a90f9fa5744d15812954fba6bb21ebf5b5b66ad78cf5816"}`, nil)

		return err
	})
}

// search a particular index for a term
func searchdb(db *containerdb.DB, metric string, term string) error {
	err := db.View(func(tx *containerdb.Tx) error {
		tx.Ascend("", func(key, value string) bool {
			contender := gjson.Get(value, metric).String()
			if strings.Contains(contender, term) {
				fmt.Printf("%s %s\n", key, contender)
			}
			return true
		})
		return nil
	})
	return err
}

// getdb search keys for a particular index for a term
func getdb(db *containerdb.DB, term string) error {
	err := db.View(func(tx *containerdb.Tx) error {
		tx.Ascend("", func(key, value string) bool {
			if strings.Contains(key, term) {
				fmt.Printf("%s %s\n", key, value)
			}
			return true
		})
		return nil
	})
	return err
}

// listdb lists all files in the databsae
func listdb(db *containerdb.DB) error {
	err := db.View(func(tx *containerdb.Tx) error {
		err := tx.Ascend("", func(key, value string) bool {
			fmt.Printf("%s\n", key)
			return true
		})
		return err
	})
	return err
}

// orderby one of the indices
func orderby(db *containerdb.DB, metric string) error {
	err := db.View(func(tx *containerdb.Tx) error {
		fmt.Printf("Order by %s\n", metric)
		err := tx.Ascend(metric, func(key, value string) bool {
			fmt.Printf("%s: %s\n", key, value)
			return true
		})
		return err
	})
	return err
}

// viewdb dumps all files and metadata
func viewdb(db *containerdb.DB) error {
	err := db.View(func(tx *containerdb.Tx) error {
		err := tx.Ascend("", func(key, value string) bool {
			fmt.Printf("%s %s\n", key, value)
			return true
		})
		return err
	})
	return err
}

func main() {

	// Search for a particular term
	searchPtr := flag.String("search", "", "Search term")

	// Get metadata for a flie
	getPtr := flag.String("get", "", "Get metadata for a particular file")

	// Index by a metric
	metricPtr := flag.String("metric", "", "Metric to index by {size|hash|name};.")

	// List files
	listPtr := flag.Bool("ls", false, "List all files.")
	flag.Parse()

	// Ensure metric is valid, if provided
	validMetrics := []string{"", "size", "sha256"}

	_, isValid := Find(validMetrics, *metricPtr)
	if !isValid {
		fmt.Printf("%s is not a valid metric\n", *metricPtr)
	} else {
		// Open an in-memory database
		db, err := containerdb.Open(":memory:")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// Initialize the database with content and indices
		initdb(db)

		// List files in database, or view all metadata
		if *listPtr == true {
			listdb(db)

			// Get metadata based on a key
		} else if *getPtr != "" {
			getdb(db, *getPtr)

			// Search for a particular term for a metric
		} else if *searchPtr != "" && *metricPtr != "" {
			searchdb(db, *metricPtr, *searchPtr)

			// Order by a specific metric
		} else if *metricPtr != "" {
			orderby(db, *metricPtr)
		} else {
			viewdb(db)
		}
	}

}

// Find takes a slice and looks for an element in it. If found it will
// Return a key if found, otherwise -1 and false
func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

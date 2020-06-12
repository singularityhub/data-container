// main.go
package main

import (
	"fmt"
	"runtime"
        "github.com/vsoch/scif-go"
)

func blockForever() {
	fmt.Println("Running sleep command...")
foo:
	runtime.Gosched()
	goto foo
}

func main() {
	blockForever()
}

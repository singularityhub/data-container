// main.go
package main

import (
	"fmt"
	"runtime"
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

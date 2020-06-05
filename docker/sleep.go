// main.go
package main

import ("runtime"
	"fmt")

func blockForever() {
    fmt.Println("Running sleep command...")
    foo: runtime.Gosched()
    goto foo
}

func main() {
    blockForever()
}

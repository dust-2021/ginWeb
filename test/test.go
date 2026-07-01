package main

import (
	"fmt"
	"sync"
)

func main() {
	var group = sync.WaitGroup{}
	for i := range 10 {
		group.Go(func() {
			fmt.Printf("number %d", i)
		})
	}
	group.Wait()
	fmt.Print("all done")
}

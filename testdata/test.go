package main

import "fmt"

func main() {
	go func() {
		fmt.Println("Test data")
	}()

	go NotCatchFn()

	go not_panik_pkg.Catch()
}

package mypkg

import "fmt"

func myfunc() {

	// False
	go func() {
		fmt.Println("Test data")
	}()

	// False
	go NotCatchFn()

	// False
	go not_panik_pkg.Catch()

	// True
	go panik.Catch()
}

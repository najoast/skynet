package util

import "fmt"

// Pcall calls a function in protected mode.
func Pcall(f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in Pcall", r)
		}
	}()
	f()
}

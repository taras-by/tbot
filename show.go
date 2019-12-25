package main

import (
	"fmt"
)

func show() (err error) {
	a := newApp()
	defer a.Close()

	for _, p := range a.storage.FindAll() {
		fmt.Printf("Participant: %v\n", p)
	}

	return nil
}

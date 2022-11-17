package main

import (
	"satoshicard/ui"
	"time"
)

func main() {
	// uictx := &ui.UIContext{}
	ui.NewUIContext()
	for {
		time.Sleep(time.Minute)
	}
}

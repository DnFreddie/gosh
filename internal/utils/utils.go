package utils

import (
	"fmt"
	"time"
)

func WaitingScreen(done chan bool, desc string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered from error:", r)
				done <- true
			}
		}()

		for {
			select {
			case <-done:
				return
			default:
				for _, r := range `-\|/` {
					fmt.Printf("\r%c %s", r, desc)
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}()
}

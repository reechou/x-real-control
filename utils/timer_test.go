package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	w := NewTimingWheel(500*time.Millisecond, 120)
	for {
		select {
		case <-w.Check(10):
			fmt.Println("10:", time.Now())
		case <-w.Check(20):
			fmt.Println("20:", time.Now())
		case <-w.Check(30):
			fmt.Println("30:", time.Now())
		case <-w.Check(60):
			fmt.Println("60:", time.Now())
		}
	}
}

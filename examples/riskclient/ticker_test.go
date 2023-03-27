package main_test

import (
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	now := time.Now()
	// now, _ := time.Parse("15:04:05.000000", "11:16:34.897123")

	nextTs := now.Round(time.Second * 30)

	if nextTs.Before(now) {
		nextTs = nextTs.Add(time.Second * 30)
	}

	t.Log(now, nextTs)

	// return

	timer := time.NewTimer(nextTs.Sub(now))
	ticker := time.NewTicker(time.Second * 30)
	ticker.Stop()

	tickCount := 0

	for {
		select {
		case ts := <-timer.C:
			t.Log("Bar timer first initialized:", ts)
			ticker.Reset(time.Second * 30)
			timer.Stop()
		case ts := <-ticker.C:
			t.Log("Bar ticker:", ts)

			tickCount++

			if tickCount > 0 {
				return
			}
		}
	}
}

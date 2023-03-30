package main

import (
	"context"
	"testing"
	"time"
)

func TestSink(t *testing.T) {
	initDB()

	t.Log(time.Minute)

	sinker, err := InsertDB[SinkAccountBar](
		context.TODO(), "operation_account_kbar",
		"TradingDay", "AccountID", "Timestamp", "Duration",
		"Open", "Close", "Highest", "Lowest",
	)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = sinker(&SinkAccountBar{
		TradingDay: "2023-03-30",
		AccountID:  "test",
		Timestamp:  time.Now().Round(time.Second).UnixMilli(),
		Duration:   time.Second.String(),
		Open:       1000,
		Close:      1100,
		Highest:    1300,
		Lowest:     980,
	}); err != nil {
		t.Fatal(err)
	}
}

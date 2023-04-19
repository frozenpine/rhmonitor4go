package client_test

import (
	"context"
	"testing"
	"time"

	"github.com/frozenpine/rhmonitor4go/service/client"
)

func TestSink(t *testing.T) {
	client.InitDB("trade.db")

	t.Log(time.Minute)

	sinker, err := client.InsertDB[client.SinkAccountBar](
		context.TODO(), "operation_account_kbar",
		"TradingDay", "AccountID", "Timestamp", "Duration",
		"Open", "Close", "Highest", "Lowest",
	)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = sinker(&client.SinkAccountBar{
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

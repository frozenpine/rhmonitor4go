package client_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/frozenpine/rhmonitor4go/service/client"
)

func TestSink(t *testing.T) {
	// if err := client.InitDB("sqlite3://trade.db"); err != nil {
	// 	t.Fatal(err)
	// }

	db, err := client.InitDB(
		fmt.Sprintf(
			"postgres://host=%s port=%d user=%s password=%s dbname=%s",
			"localhost", 5432, "trade", "trade", "lingma",
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// sinker, err := client.InsertDB[client.SinkAccountBar](
	// 	context.TODO(), "trade.operation_account_kbar",
	// 	"TradingDay", "AccountID", "Timestamp", "Duration",
	// 	"Open", "Close", "Highest", "Lowest",
	// )
	sinker, err := client.InsertDB[client.SinkAccountBar](
		context.TODO(), "operation_account_kbar",
	)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()

	if result, err := sinker(&client.SinkAccountBar{
		TradingDay: now.Format("2006-01-02"),
		AccountID:  "test",
		Timestamp:  now.Round(time.Millisecond),
		Duration:   time.Second.String(),
		Open:       float64(now.Hour()),
		Close:      float64(now.UnixMilli()),
		Highest:    float64(now.Minute()),
		Lowest:     float64(now.Second()),
	}); err != nil {
		t.Fatal(err)
	} else {
		id, _ := result.LastInsertId()
		count, _ := result.RowsAffected()

		t.Logf("rowid: %d, insert count: %d", id, count)
	}
}

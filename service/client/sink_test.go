package client_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/frozenpine/rhmonitor4go/service/client"
	"github.com/vmihailenco/msgpack/v5"
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

func TestMarshal(t *testing.T) {
	now := time.Now()

	acct := client.SinkAccount{
		InvestorID: "test",
		TradingDay: now.Format("2006-01-02"),
		Timestamp:  now,
		PreBalance: 10000,
		Balance:    10000,
		Deposit:    0,
		Withdraw:   0,
		Profit:     0,
		Fee:        0,
		Margin:     0,
		Available:  10000,
	}

	buffer, err := msgpack.Marshal(acct)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(buffer)
	}

	v := client.SinkAccount{}
	if err = msgpack.Unmarshal(buffer, &v); err != nil {
		t.Fatal(err)
	}
	t.Log(v)

	payload := "\x8b\xaaaccount_id\xa7default\xabtrading_day\xa820230426\xa9timestamp\xd7\xff\x02\x83\xb0`dH\xb3\x90\xabpre_balance\xcb\x00\x00\x00\x00\x00\x00\x00\x00\xa7balance\xcb\x00\x00\x00\x00\x00\x00\x00\x00\xa7deposit\xcb\x00\x00\x00\x00\x00\x00\x00\x00\xa8withdraw\xcb\x00\x00\x00\x00\x00\x00\x00\x00\xa6profit\xcb\x00\x00\x00\x00\x00\x00\x00\x00\xa3fee\xcb\x00\x00\x00\x00\x00\x00\x00\x00\xa6margin\xcb\x00\x00\x00\x00\x00\x00\x00\x00\xa9available\xcb\x00\x00\x00\x00\x00\x00\x00\x00"

	if err = msgpack.Unmarshal([]byte(payload), &v); err != nil {
		t.Fatal(err)
	}
	t.Log(v)
}

func TestDuration(t *testing.T) {
	multiple := []int{1, 5, 24, 30}

	for _, dur := range []time.Duration{time.Minute, time.Hour} {
		for _, v := range multiple {
			t.Log(time.Duration(v) * dur)
		}
	}

	t.Log(time.ParseDuration("24h"))
	t.Log(time.Hour * 24)
}

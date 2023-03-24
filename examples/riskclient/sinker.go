package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"unsafe"

	_ "github.com/mattn/go-sqlite3"

	"github.com/frozenpine/rhmonitor4go/service"
)

const (
	structuresSQL = `
-- ----------------------------
-- Table structure for operation_account_kbar
-- ----------------------------
CREATE TABLE IF NOT EXISTS "operation_account_kbar" (
"trading_day" DATE NOT NULL,
"account_id" VARCHAR NOT NULL,
"timestamp" TIMESTAMP NOT NULL,
"duration" VARCHAR(255) NOT NULL,
"open" FLOAT NOT NULL,
"high" FLOAT NOT NULL,
"low" FLOAT NOT NULL,
"close" FLOAT NOT NULL,
PRIMARY KEY ("trading_day", "account_id", "timestamp", "duration")
);

-- ----------------------------
-- Table structure for operation_trading_account
-- ----------------------------
CREATE TABLE IF NOT EXISTS "operation_trading_account" (
"trading_day" DATE NOT NULL,
"account_id" VARCHAR NOT NULL,
"timestamp" TIMESTAMP NOT NULL,
"pre_balance" FLOAT NOT NULL,
"balance" FLOAT NOT NULL,
"deposit" FLOAT NOT NULL,
"withdraw" FLOAT NOT NULL,
"profit" FLOAT NOT NULL,
"fee" FLOAT NOT NULL,
"margin" FLOAT NOT NULL,
"available" FLOAT NOT NULL,
PRIMARY KEY ("trading_day", "account_id", "timestamp")
);`
)

type SinkAccount struct {
	InvestorID string
	TradingDay string
	Timestamp  int64
	PreBalance float64
	Balance    float64
	Deposit    float64
	Withdraw   float64
	Profit     float64
	Fee        float64
	Margin     float64
	Available  float64
}

func (acct *SinkAccount) FromAccount(value *service.Account) {
	acct.TradingDay = value.TradingDay
	acct.InvestorID = value.Investor.InvestorId
	acct.Timestamp = value.Timestamp
	acct.PreBalance = value.PreBalance
	acct.Deposit = value.Deposit
	acct.Withdraw = value.Withdraw
	acct.Profit = value.CloseProfit + value.PositionProfit
	acct.Fee = value.Commission + value.FrozenCommission
	acct.Margin = value.CurrentMargin + value.FrozenMargin
	acct.Available = value.Available
	acct.Balance = acct.PreBalance + acct.Deposit - acct.Withdraw + acct.Profit - value.Commission
}

func (acct *SinkAccount) Compare(other *SinkAccount) (diff bool) {
	if acct.InvestorID != "" && acct.InvestorID != other.InvestorID {
		log.Print("Compare to diff investors:", acct.InvestorID, other.InvestorID)
		return
	}

	if acct.Timestamp >= other.Timestamp {
		log.Printf("Compare to old tick: \n%+v\n%+v", acct, other)
		return
	}

	if acct.TradingDay != other.TradingDay {
		acct.TradingDay = other.TradingDay
		diff = true
	}

	if acct.PreBalance != other.PreBalance {
		acct.PreBalance = other.PreBalance
		diff = true
	}

	if acct.Deposit != other.Deposit {
		acct.Deposit = other.Deposit
		diff = true
	}

	if acct.Withdraw != other.Withdraw {
		acct.Withdraw = other.Withdraw
		diff = true
	}

	if acct.Profit != other.Profit {
		acct.Profit = other.Profit
		diff = true
	}

	if acct.Fee != other.Fee {
		acct.Fee = other.Fee
		diff = true
	}

	if acct.Margin != other.Margin {
		acct.Margin = other.Margin
		diff = true
	}

	if acct.Available != other.Available {
		acct.Available = other.Available
		diff = true
	}

	if acct.Balance != other.Balance {
		acct.Balance = other.Balance
		diff = true
	}

	return
}

var (
	db *sql.DB
)

func initDB() (err error) {
	if db, err = sql.Open("sqlite3", dbFile); err != nil {
		log.Fatal("Open database failed:", err)
	}

	if _, err = db.Exec(structuresSQL); err != nil {
		log.Fatal("Create table failed:", err)
	}

	return
}

func insertDB[T any](
	ctx context.Context,
	query string,
	argNames ...string,
) (func(v interface{}) (sql.Result, error), error) {
	argCount := strings.Count(query, "?")
	if len(argNames) != argCount {
		return nil, errors.New("args mismatch in query")
	}

	typ := reflect.TypeOf(new(T)).Elem()

	fieldOffsets := []struct {
		offset uintptr
		typ    reflect.Type
	}{}

	for _, name := range argNames {
		if field, ok := typ.FieldByName(name); !ok {
			return nil, fmt.Errorf("%s has no field name: %s", typ.Name(), name)
		} else {
			fieldOffsets = append(fieldOffsets, struct {
				offset uintptr
				typ    reflect.Type
			}{
				offset: field.Offset,
				typ:    field.Type,
			})
		}
	}

	getArgList := func(v interface{}) []interface{} {
		basePtr := reflect.Indirect((reflect.ValueOf(v))).Addr().Pointer()

		argList := []interface{}{}

		for _, field := range fieldOffsets {
			argList = append(
				argList,
				reflect.Indirect(reflect.NewAt(
					field.typ, unsafe.Pointer(basePtr+field.offset),
				)).Interface(),
			)
		}

		return argList
	}

	return func(v interface{}) (sql.Result, error) {
		return db.ExecContext(ctx, query, getArgList(v)...)
	}, nil
}

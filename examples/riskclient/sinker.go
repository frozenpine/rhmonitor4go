package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
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

var (
	ErrSameData = errors.New("same with old data")
)

type SinkAccount struct {
	InvestorID string  `sql:"account_id"`
	TradingDay string  `sql:"trading_day"`
	Timestamp  int64   `sql:"timestamp"`
	PreBalance float64 `sql:"pre_balance"`
	Balance    float64 `sql:"balance"`
	Deposit    float64 `sql:"deposit"`
	Withdraw   float64 `sql:"withdraw"`
	Profit     float64 `sql:"profit"`
	Fee        float64 `sql:"fee"`
	Margin     float64 `sql:"margin"`
	Available  float64 `sql:"available"`
}

func (acct *SinkAccount) FromAccount(value *service.Account) {
	acct.TradingDay = value.TradingDay
	acct.InvestorID = value.Investor.InvestorId
	acct.Timestamp = time.Now().UnixMilli()
	acct.PreBalance = value.PreBalance
	acct.Deposit = value.Deposit
	acct.Withdraw = value.Withdraw
	acct.Profit = value.CloseProfit + value.PositionProfit
	acct.Fee = value.Commission + value.FrozenCommission
	acct.Margin = value.CurrentMargin + value.FrozenMargin
	acct.Available = value.Available
	acct.Balance = acct.PreBalance + acct.Deposit - acct.Withdraw + acct.Profit - value.Commission
}

type BarMode uint8

const (
	Continuous BarMode = 1 << iota
	FirstTick
)

type SinkAccountBar struct {
	TradingDay string  `sql:"trading_day"`
	AccountID  string  `sql:"account_id"`
	Timestamp  int64   `sql:"timestamp"`
	Duration   string  `sql:"duration"`
	Open       float64 `sql:"open"`
	Close      float64 `sql:"close"`
	Highest    float64 `sql:"high"`
	Lowest     float64 `sql:"low"`
}

type AccountSinker struct {
	mode       BarMode
	barSinker  func(*SinkAccountBar) (sql.Result, error)
	acctSinker func(*SinkAccount) (sql.Result, error)

	ctx         context.Context
	source      service.RohonMonitor_SubInvestorMoneyClient
	output      chan *service.Account
	waterMark   *service.Account
	accountPool sync.Pool
	barPool     sync.Pool

	duration     time.Duration
	settlements  map[string]*service.Account
	tradingDay   string
	accountCache map[string][]*service.Account
	barCache     map[string]*SinkAccountBar
}

func NewAccountSinker(
	ctx context.Context,
	mode BarMode, dur time.Duration,
	settlements map[string]*service.Account,
	src service.RohonMonitor_SubInvestorMoneyClient,
) (*AccountSinker, error) {
	if src == nil || settlements == nil {
		return nil, errors.New("invalid sinker args")
	}

	acctSinker, err := InsertDB[SinkAccount](
		ctx, "operation_trading_account",
		"TradingDay", "InvestorID", "Timestamp", "PreBalance",
		"Balance", "Deposit", "Withdraw", "Profit", "Fee",
		"Margin", "Available",
	)
	if err != nil {
		return nil, err
	}

	barSinker, err := InsertDB[SinkAccountBar](
		ctx, "operation_account_kbar",
		"TradingDay", "AccountID", "Timestamp", "Duration",
		"Open", "Close", "Highest", "Lowest",
	)
	if err != nil {
		return nil, err
	}

	tradingDay := ""
	for _, v := range settlements {
		tradingDay = v.GetTradingDay()
		break
	}

	sinker := &AccountSinker{
		ctx:        ctx,
		mode:       mode,
		acctSinker: acctSinker,
		barSinker:  barSinker,
		source:     src,
		output:     make(chan *service.Account, 1),
		waterMark: &service.Account{
			TradingDay: tradingDay,
			Investor: &service.Investor{
				InvestorId: "default",
			},
		},
		duration:     dur,
		accountPool:  sync.Pool{New: func() any { return new(SinkAccount) }},
		barPool:      sync.Pool{New: func() any { return new(SinkAccountBar) }},
		tradingDay:   tradingDay,
		settlements:  settlements,
		accountCache: make(map[string][]*service.Account),
		barCache:     make(map[string]*SinkAccountBar),
	}

	go sinker.run()

	return sinker, nil
}

func (sink *AccountSinker) newSinkAccount() *SinkAccount {
	data := sink.accountPool.Get()
	runtime.SetFinalizer(data, sink.accountPool.Put)
	return data.(*SinkAccount)
}

func (sink *AccountSinker) newSinkBar() *SinkAccountBar {
	data := sink.barPool.Get()
	runtime.SetFinalizer(data, sink.barPool.Put)
	return data.(*SinkAccountBar)
}

func (sink *AccountSinker) boundary(ts time.Time) {
	sink.waterMark.Timestamp = ts.UnixMilli()

	for accountID, settAccount := range sink.settlements {
		accountList := sink.accountCache[accountID]

		preBar := sink.barCache[accountID]

		if preBar == nil {
			preBar = &SinkAccountBar{
				AccountID:  accountID,
				TradingDay: sink.tradingDay,
				Close:      settAccount.PreBalance,
			}
			sink.barCache[accountID] = preBar
		}

		var (
			open             = preBar.Close
			high, low, close float64
		)

		for _, v := range accountList {
			if high == 0 {
				high = v.Balance
			} else if v.Balance > high {
				high = v.Balance
			}

			if low == 0 {
				low = v.Balance
			} else if v.Balance < low {
				low = v.Balance
			}
		}

		count := len(accountList)

		if count > 0 {
			if sink.mode == FirstTick {
				open = accountList[0].Balance
			}

			close = accountList[count-1].Balance
		} else {
			close = open
		}

		bar := sink.newSinkBar()
		bar.TradingDay = sink.tradingDay
		bar.AccountID = accountID
		bar.Timestamp = ts.Round(sink.duration).UnixMilli()
		bar.Duration = sink.duration.String()
		bar.Open = open
		bar.Close = close
		bar.Highest = high
		bar.Lowest = low

		if _, err := sink.barSinker(bar); err != nil {
			log.Printf("Sink account bar failed: %+v", err)
		}

		sink.barCache[accountID] = bar
	}

	sink.output <- sink.waterMark
}

func (sink *AccountSinker) run() {
	inputChan := make(chan *service.Account, 1)

	go func() {
		log.Print("Starting gRPC Account data receiver")
		for {
			acct, err := sink.source.Recv()

			if err != nil {
				log.Printf("Receive investor's account failed: %+v", err)
				break
			}

			fmt.Printf("OnRtnInvestorMoney %+v\n", acct)
			inputChan <- acct
		}
	}()

	now := time.Now()
	nextTs := now.Round(sink.duration)

	if nextTs.Before(now) {
		nextTs = nextTs.Add(sink.duration)
	}

	timer := time.NewTimer(nextTs.Sub(now))
	ticker := time.NewTicker(sink.duration)

	ticker.Stop()

	for {
		select {
		case <-sink.ctx.Done():
			return
		case ts := <-timer.C:
			log.Printf("Bar ticker first initialized: %+v", ts)
			ticker.Reset(sink.duration)
			timer.Stop()

			sink.boundary(ts)
		case ts := <-ticker.C:
			sink.boundary(ts)
		case acct := <-inputChan:
			sinkAccount := sink.newSinkAccount()
			sinkAccount.FromAccount(acct)

			if _, err := sink.acctSinker(sinkAccount); err != nil {
				log.Print("Sink data failed:", err)
			}

			sink.accountCache[acct.Investor.InvestorId] = append(
				sink.accountCache[acct.Investor.InvestorId],
				acct,
			)

			sink.output <- acct
		}
	}
}

func (sink *AccountSinker) Data() <-chan *service.Account {
	return sink.output
}

var (
	db *sql.DB
)

func initDB() (err error) {
	log.Print("Try to open db file:", dbFile)

	if db, err = sql.Open("sqlite3", dbFile); err != nil {
		log.Fatal("Open database failed:", dbFile, err)
	}

	if _, err = db.Exec(structuresSQL); err != nil {
		log.Fatalf(
			"Create table failed: %s, %s\n%s",
			dbFile, err, structuresSQL,
		)
	}

	return
}

type fieldOffset struct {
	offset uintptr
	typ    reflect.Type
}

func InsertDB[T any](
	ctx context.Context,
	tblName string,
	argNames ...string,
) (func(v *T) (sql.Result, error), error) {
	if tblName == "" {
		return nil, errors.New("no table name")
	}

	obj := new(T)
	typ := reflect.TypeOf(obj).Elem()

	argLen := len(argNames)

	fieldOffsets := make([]fieldOffset, argLen)
	sqlFields := make([]string, argLen)
	argList := make([]string, argLen)

	if argLen == 0 {
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			sqlField := field.Tag.Get("sql")

			if sqlField == "" {
				continue
			}

			fieldOffsets = append(fieldOffsets, fieldOffset{
				offset: field.Offset,
				typ:    field.Type,
			})

			sqlFields = append(sqlFields, sqlField)
			argList = append(argList, "?")
		}
	} else {
		for idx, name := range argNames {
			if field, ok := typ.FieldByName(name); !ok {
				return nil, fmt.Errorf("%s has no field name: %s", typ.Name(), name)
			} else {
				sqlField := field.Tag.Get("sql")
				if sqlField == "" {
					return nil, fmt.Errorf("%s has no sql tag", field.Name)
				}

				fieldOffsets[idx] = fieldOffset{
					offset: field.Offset,
					typ:    field.Type,
				}

				sqlFields[idx] = sqlField
				argList[idx] = "?"
			}
		}
	}

	sqlTpl := fmt.Sprintf(
		"INSERT INTO %s(%s) VALUES (%s);",
		tblName,
		strings.Join(sqlFields, ","),
		strings.Join(argList, ","),
	)

	getArgList := func(v interface{}) []interface{} {
		basePtr := reflect.Indirect(reflect.ValueOf(v)).Addr().Pointer()

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

	return func(v *T) (sql.Result, error) {
		values := getArgList(v)

		log.Printf("Executing: %s, %s, %+v", sqlTpl, values, v)

		return db.ExecContext(ctx, sqlTpl, values...)
	}, nil
}

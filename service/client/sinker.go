package client

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"

	_ "github.com/jackc/pgx/v5/stdlib"
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

	driverMap = map[string]string{
		"postgres": "pgx",
	}
)

type MsgFormat uint8

func (msgFmt *MsgFormat) Set(value string) error {
	switch value {
	case "proto3":
		*msgFmt = MsgProto3
	case "json":
		*msgFmt = MsgJson
	case "msgpack":
		*msgFmt = MsgPack
	default:
		return errors.New("invalid msg format")
	}

	return nil
}

//go:generate stringer -type MsgFormat -linecomment
const (
	MsgProto3 MsgFormat = iota // proto3
	MsgJson                    // json
	MsgPack                    // msgpack
)

type SinkAccount struct {
	InvestorID string    `sql:"account_id" json:"account_id" msgpack:"account_id"`
	TradingDay string    `sql:"trading_day" json:"trading_day" msgpack:"trading_day"`
	Timestamp  time.Time `sql:"timestamp" json:"timestamp" msgpack:"timestamp"`
	PreBalance float64   `sql:"pre_balance" json:"pre_balance" msgpack:"pre_balance"`
	Balance    float64   `sql:"balance" json:"balance" msgpack:"balance"`
	Deposit    float64   `sql:"deposit" json:"deposit" msgpack:"deposit"`
	Withdraw   float64   `sql:"withdraw" json:"withdraw" msgpack:"withdraw"`
	Profit     float64   `sql:"profit" json:"profit" msgpack:"profit"`
	Fee        float64   `sql:"fee" json:"fee" msgpack:"fee"`
	Margin     float64   `sql:"margin" json:"margin" msgpack:"margin"`
	Available  float64   `sql:"available" json:"available" msgpack:"available"`
}

func (acct *SinkAccount) FromAccount(value *service.Account) {
	acct.TradingDay = value.TradingDay
	acct.InvestorID = value.Investor.InvestorId
	acct.Timestamp = time.UnixMilli(value.Timestamp)
	acct.PreBalance = value.PreBalance
	acct.Deposit = value.Deposit
	acct.Withdraw = value.Withdraw
	acct.Profit = value.CloseProfit + value.PositionProfit
	acct.Fee = value.Commission + value.FrozenCommission
	acct.Margin = value.CurrentMargin + value.FrozenMargin
	acct.Available = value.Available
	acct.Balance = service.RohonCaculateDynamicBalance(
		value.PreBalance, value.Deposit, value.Withdraw,
		value.CloseProfit, value.PositionProfit, value.Commission,
	)
}

type BarMode uint8

const (
	Continuous BarMode = 1 << iota
	FirstTick
)

type SinkAccountBar struct {
	TradingDay string    `sql:"trading_day" json:"trading_day" msgpack:"trading_day"`
	AccountID  string    `sql:"account_id" json:"account_id" msgpack:"account_id"`
	Timestamp  time.Time `sql:"timestamp" json:"timestamp" msgpack:"timestamp"`
	Duration   string    `sql:"duration" json:"duration" msgpack:"duration"`
	Open       float64   `sql:"open" json:"open" msgpack:"open"`
	Close      float64   `sql:"close" json:"close" msgpack:"close"`
	Highest    float64   `sql:"high" json:"high" msgpack:"high"`
	Lowest     float64   `sql:"low" json:"low" msgpack:"low"`
}

type AccountSinker struct {
	mode       BarMode
	barSinker  func(*SinkAccountBar) (sql.Result, error)
	acctSinker func(*SinkAccount) (sql.Result, error)

	ctx         context.Context
	source      service.RohonMonitor_SubInvestorMoneyClient
	output      chan *SinkAccount
	waterMark   *SinkAccount
	accountPool sync.Pool
	barPool     sync.Pool

	duration     time.Duration
	settlements  map[string]*service.Account
	tradingDay   string
	accountCache map[string][]*SinkAccount
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
		output:     make(chan *SinkAccount, 1),
		waterMark: &SinkAccount{
			TradingDay: tradingDay,
			InvestorID: "default",
		},
		duration:     dur,
		accountPool:  sync.Pool{New: func() any { return new(SinkAccount) }},
		barPool:      sync.Pool{New: func() any { return new(SinkAccountBar) }},
		tradingDay:   tradingDay,
		settlements:  settlements,
		accountCache: make(map[string][]*SinkAccount),
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
	sink.waterMark.Timestamp = ts

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
			open, high, low, close float64
			count                  = len(accountList)
		)

		if sink.mode == FirstTick && count > 0 {
			open, high, low, close = accountList[0].Balance, accountList[0].Balance, accountList[0].Balance, accountList[0].Balance
		} else {
			open, high, low, close = preBar.Close, preBar.Close, preBar.Close, preBar.Close
		}

		for idx, v := range accountList {
			if v.Balance > high {
				high = v.Balance
			}

			if v.Balance < low {
				low = v.Balance
			}

			if idx == count-1 {
				close = v.Balance
			}
		}

		bar := sink.newSinkBar()
		bar.TradingDay = sink.tradingDay
		bar.AccountID = accountID
		bar.Timestamp = ts.Round(sink.duration)
		bar.Duration = sink.duration.String()
		bar.Open = open
		bar.Close = close
		bar.Highest = high
		bar.Lowest = low

		if _, err := sink.barSinker(bar); err != nil {
			log.Printf("Sink account bar failed: %+v", err)
		}

		sink.barCache[accountID] = bar
		sink.accountCache[accountID] = []*SinkAccount{}
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

			sink.accountCache[sinkAccount.InvestorID] = append(
				sink.accountCache[sinkAccount.InvestorID],
				sinkAccount,
			)

			sink.output <- sinkAccount
		}
	}
}

func (sink *AccountSinker) Data() <-chan *SinkAccount {
	return sink.output
}

var (
	db          *sql.DB
	connPattern = regexp.MustCompile(
		"(?P<proto>(?:sqlite3|postgres|pgx))://(?P<value>.+)",
	)
)

func InitDB(conn string) (c *sql.DB, err error) {
	matchs := connPattern.FindStringSubmatch(conn)
	if len(matchs) < 1 {
		return nil, errors.New("invalid db conn string")
	}

	protoIdx := connPattern.SubexpIndex("proto")
	valueIdx := connPattern.SubexpIndex("value")

	var (
		proto string
		exist bool
	)

	if proto, exist = driverMap[matchs[protoIdx]]; !exist {
		proto = matchs[protoIdx]
	}

	log.Print("Try to open db: ", conn)

	if c, err = sql.Open(proto, matchs[valueIdx]); err != nil {
		log.Fatalf("Parse database[%s] failed: %+v", conn, err)
	} else if err = c.Ping(); err != nil {
		log.Fatalf("Open database[%s] failed: %+v", conn, err)
	}

	if _, err = c.Exec(structuresSQL); err != nil {
		log.Fatalf(
			"Create table failed: %s, %s\n%s",
			conn, err, structuresSQL,
		)
	}

	db = c

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
			argList = append(argList, fmt.Sprintf("$%d", i+1))
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
				argList[idx] = fmt.Sprintf("$%d", idx+1)
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

	// tx, err := db.Begin()

	// insStmt, err := tx.Prepare(sqlTpl)
	// if err != nil {
	// 	log.Fatalf("Prepare sql \"%s\" failed: %+v", sqlTpl, err)
	// }

	return func(v *T) (sql.Result, error) {
		values := getArgList(v)
		// defer tx.Commit()

		// log.Printf("Executing: %s, %s, %+v", sqlTpl, values, v)

		c, cancel := context.WithCancel(ctx)
		defer cancel()

		return db.ExecContext(c, sqlTpl, values...)
		// return insStmt.ExecContext(ctx, values...)
	}, nil
}

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
	"sync/atomic"
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
	driverMap = map[string]string{
		"postgres": "pgx",
	}
)

type AccountSinker struct {
	mode       BarMode
	barSinker  func(*SinkAccountBar) (sql.Result, error)
	acctSinker func(*SinkAccount) (sql.Result, error)

	ctx         context.Context
	source      service.RohonMonitor_SubInvestorMoneyClient
	output      chan *SinkAccountBar
	accountPool sync.Pool
	barPool     sync.Pool

	duration     time.Duration
	settlements  sync.Map
	boundaryFlag atomic.Bool
	barCache     map[string]*SinkAccountBar
}

func NewAccountSinker(
	ctx context.Context,
	mode BarMode, dur time.Duration,
	settlements []*service.Account,
	src service.RohonMonitor_SubInvestorMoneyClient,
) (*AccountSinker, error) {
	if src == nil || settlements == nil || len(settlements) == 0 {
		return nil, errors.New("invalid sinker args")
	}

	acctSinker, err := InsertDB[SinkAccount](
		ctx, "operation_trading_account",
		// "TradingDay", "InvestorID", "Timestamp", "PreBalance",
		// "Balance", "Deposit", "Withdraw", "Profit", "Fee",
		// "Margin", "Available",
	)
	if err != nil {
		return nil, err
	}

	barSinker, err := InsertDB[SinkAccountBar](
		ctx, "operation_account_kbar",
		// "TradingDay", "AccountID", "Timestamp", "Duration",
		// "Open", "Close", "Highest", "Lowest",
	)
	if err != nil {
		return nil, err
	}

	sinker := &AccountSinker{
		ctx:         ctx,
		mode:        mode,
		acctSinker:  acctSinker,
		barSinker:   barSinker,
		source:      src,
		output:      make(chan *SinkAccountBar, 1),
		duration:    dur,
		accountPool: sync.Pool{New: func() any { return new(SinkAccount) }},
		barPool:     sync.Pool{New: func() any { return new(SinkAccountBar) }},
	}

	for _, settle := range settlements {
		investor := settle.GetInvestor()

		if investor == nil {
			log.Printf("Investor info not found in settle: %+v", settle)
			continue
		}

		sinker.settlements.Store(investor.InvestorId, settle)
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

// boundary 仅处理无实时账户流更新的bar数据
func (sink *AccountSinker) boundary(ts time.Time) {
	if !sink.boundaryFlag.Load() {
		log.Print("No account stream input, stop bar boundary")
		return
	}

	sink.settlements.Range(func(key, value any) bool {
		accountID := key.(string)
		settAccount := value.(*service.Account)
		boundaryTs := ts.Round(sink.duration)

		currBar, exist := sink.barCache[accountID]
		if !exist {
			currBar = &SinkAccountBar{
				TradingDay: settAccount.TradingDay,
				AccountID:  accountID,
				Duration:   sink.duration.String(),
				Open:       settAccount.PreBalance,
				Highest:    settAccount.PreBalance,
				Lowest:     settAccount.PreBalance,
				Close:      settAccount.PreBalance,
			}

			sink.barCache[accountID] = currBar
		} else if currBar.Timestamp.After(boundaryTs) {
			// 已由实时流触发更新bar数据
			return true
		}

		currBar.Timestamp = boundaryTs

		if _, err := sink.barSinker(currBar); err != nil {
			log.Printf("Sink account bar failed: %+v", err)
		}

		sink.output <- currBar

		return true
	})
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
	streamTimeout := time.NewTimer(time.Second * 2)

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
		case <-streamTimeout.C:
			sink.boundaryFlag.Store(false)
		case acct := <-inputChan:
			sink.boundaryFlag.Store(true)
			streamTimeout.Stop()
			streamTimeout.Reset(time.Second * 2)

			sinkAccount := sink.newSinkAccount()
			sinkAccount.FromAccount(acct)

			if _, err := sink.acctSinker(sinkAccount); err != nil {
				log.Print("Sink data failed:", err)
			}

			currBar, exist := sink.barCache[sinkAccount.InvestorID]
			if !exist {
				ts := sinkAccount.Timestamp.Round(sink.duration)
				if ts.Before(sinkAccount.Timestamp) {
					ts = ts.Add(sink.duration)
				}

				var price float64

				if sett, exist := sink.settlements.Load(sinkAccount.InvestorID); exist {
					// 昨结算账户存在
					price = sett.(*service.Account).PreBalance
				} else {
					// 昨结算账户不存在，可能为实时上场新增账户，以账户 昨结算 + 入金 - 出金 作为初始资金
					price = sinkAccount.PreBalance + sinkAccount.Deposit - sinkAccount.Withdraw
				}

				currBar = sink.newSinkBar()
				currBar.TradingDay = sinkAccount.TradingDay
				currBar.AccountID = sinkAccount.InvestorID
				currBar.Timestamp = ts
				currBar.Duration = sink.duration.String()
				currBar.Open = price
				currBar.Highest = price
				currBar.Lowest = price
				currBar.Close = price
			}

			// bar切换
			if sinkAccount.Timestamp.After(currBar.Timestamp) {
				preBar := currBar

				if _, err := sink.barSinker(preBar); err != nil {
					log.Printf("Sink account bar failed: %+v", err)
				}

				currBar = sink.newSinkBar()
				currBar.TradingDay = sinkAccount.TradingDay
				ts := preBar.Timestamp.Add(sink.duration)
				if sinkAccount.Timestamp.After(ts) {
					if ts = sinkAccount.Timestamp.Round(sink.duration); sinkAccount.Timestamp.After(ts) {
						ts = ts.Add(sink.duration)
					}
				}
				currBar.Timestamp = ts
				currBar.AccountID = sinkAccount.InvestorID
				currBar.Duration = sink.duration.String()

				if sink.mode == FirstTick {
					currBar.Open = sinkAccount.Balance
					currBar.Highest = sinkAccount.Balance
					currBar.Lowest = sinkAccount.Balance
					currBar.Close = sinkAccount.Balance
				} else {
					currBar.Open = preBar.Close
					currBar.Highest = preBar.Close
					currBar.Lowest = preBar.Close
					currBar.Close = preBar.Close
				}
			}

			if sinkAccount.Balance > currBar.Highest {
				currBar.Highest = sinkAccount.Balance
			} else if sinkAccount.Balance < currBar.Lowest {
				currBar.Lowest = sinkAccount.Balance
			}

			currBar.Close = sinkAccount.Balance

			sink.barCache[sinkAccount.InvestorID] = currBar

			sink.output <- currBar
		}
	}
}

func (sink *AccountSinker) Data() <-chan *SinkAccountBar {
	return sink.output
}

func (sink *AccountSinker) RenewSettle(acct *service.Account) error {
	if acct == nil {
		return errors.New("invalid settlement account")
	}

	investor := acct.GetInvestor()
	if investor == nil {
		return errors.New("no investor info found")
	}

	sink.settlements.Store(investor.InvestorId, acct)

	return nil
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

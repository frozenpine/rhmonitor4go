package sinker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/frozenpine/rhmonitor4go/service"
)

// const (
// 	acctStructuresSQL = `
// -- ----------------------------
// -- Table structure for operation_account_kbar
// -- ----------------------------
// CREATE TABLE IF NOT EXISTS "operation_account_kbar" (
// "trading_day" DATE NOT NULL,
// "account_id" VARCHAR NOT NULL,
// "timestamp" TIMESTAMPTZ	 NOT NULL,
// "duration" VARCHAR(255) NOT NULL,
// "open" FLOAT NOT NULL,
// "high" FLOAT NOT NULL,
// "low" FLOAT NOT NULL,
// "close" FLOAT NOT NULL,
// PRIMARY KEY ("trading_day", "account_id", "timestamp", "duration")
// ) PARTITION BY RANGE(trading_day);

// -- ----------------------------
// -- Table structure for operation_trading_account
// -- ----------------------------
// CREATE TABLE IF NOT EXISTS "operation_trading_account" (
// "trading_day" DATE NOT NULL,
// "account_id" VARCHAR NOT NULL,
// "timestamp" TIMESTAMPTZ NOT NULL,
// "pre_balance" FLOAT NOT NULL,
// "balance" FLOAT NOT NULL,
// "deposit" FLOAT NOT NULL,
// "withdraw" FLOAT NOT NULL,
// "profit" FLOAT NOT NULL,
// "fee" FLOAT NOT NULL,
// "margin" FLOAT NOT NULL,
// "available" FLOAT NOT NULL,
// PRIMARY KEY ("trading_day", "account_id", "timestamp")
// ) PARTITION BY RANGE(trading_day);`
// )

func roundTS(ts time.Time, duration time.Duration, up bool) time.Time {
	result := ts.Round(duration)

	if up {
		if result.Before(ts) {
			return result.Add(duration)
		}
	} else {
		if result.After(ts) {
			return result.Add(-duration)
		}
	}

	return result
}

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
	barTs        time.Time
	barCache     map[string]*SinkAccountBar
	accountCache map[string]*SinkAccount
}

func NewAccountSinker(
	ctx context.Context,
	mode BarMode, dur time.Duration,
	settlements []*service.Account,
	src service.RohonMonitor_SubInvestorMoneyClient,
) (*AccountSinker, error) {
	if src == nil || settlements == nil || len(settlements) == 0 {
		return nil, ErrInvalidArgs
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
		ctx:          ctx,
		mode:         mode,
		acctSinker:   acctSinker,
		barSinker:    barSinker,
		source:       src,
		output:       make(chan *SinkAccountBar, 1),
		duration:     dur,
		accountPool:  sync.Pool{New: func() any { return new(SinkAccount) }},
		barPool:      sync.Pool{New: func() any { return new(SinkAccountBar) }},
		barTs:        roundTS(time.Now(), dur, true),
		barCache:     make(map[string]*SinkAccountBar),
		accountCache: make(map[string]*SinkAccount),
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

// boundary 全部结算账号的bar数据入库
func (sink *AccountSinker) boundary(boundaryTs time.Time) {
	preTs := boundaryTs.Add(-sink.duration)

	sink.settlements.Range(func(key, value any) bool {
		accountID := key.(string)
		settAccount := value.(*service.Account)

		currBar, exist := sink.barCache[accountID]
		if !exist {
			// boundary触发时无资金信息更新的账户
			currBar = sink.newSinkBar()

			currBar.TradingDay = settAccount.TradingDay
			currBar.AccountID = accountID
			currBar.Duration = sink.duration.String()
			currBar.Open = settAccount.PreBalance
			currBar.Highest = settAccount.PreBalance
			currBar.Lowest = settAccount.PreBalance
			currBar.Close = settAccount.PreBalance

			sink.barCache[accountID] = currBar
		}

		currBar.TradingDay = settAccount.TradingDay
		currBar.Timestamp = boundaryTs
		// 当前bar时间范围无资金流更新或不存在资金流数据
		if acct, exist := sink.accountCache[accountID]; !exist ||
			!acct.Timestamp.After(preTs) {
			currBar.Highest = currBar.Open
			currBar.Lowest = currBar.Open
			currBar.Close = currBar.Open

			sink.output <- currBar
		}

		if _, err := sink.barSinker(currBar); err != nil {
			log.Printf("Sink bar data failed: %+v", err)
		}

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
	streamTimeout := time.NewTimer(sink.duration)
	streamTimeout.Stop()

	for {
		select {
		case <-sink.ctx.Done():
			return
		case ts := <-streamTimeout.C:
			log.Printf("Account stream timeout: %+v", ts)
			sink.boundary(roundTS(ts, sink.duration, false))
		case acct := <-inputChan:
			sinkAccount := sink.newSinkAccount()
			sinkAccount.FromAccount(acct)

			if _, err := sink.acctSinker(sinkAccount); err != nil {
				log.Print("Sink account data failed:", err)
			}
			sink.accountCache[sinkAccount.InvestorID] = sinkAccount

			if sinkAccount.Timestamp.After(sink.barTs) {
				// 由资金流时间戳驱动全局bar boundary
				nextTs := roundTS(sinkAccount.Timestamp, sink.duration, true)

				if sink.barTs.Add(sink.duration).Equal(nextTs) {
					sink.boundary(sink.barTs)

					streamTimeout.Stop()
					streamTimeout.Reset(sink.duration)
				}

				sink.barTs = nextTs
			}

			currBar, exist := sink.barCache[sinkAccount.InvestorID]
			if !exist {
				// 程序启动账号的首笔资金更新
				var preBalance float64

				if sett, exist := sink.settlements.Load(sinkAccount.InvestorID); exist {
					// 昨结算账户存在
					preBalance = sett.(*service.Account).PreBalance
				} else {
					// 昨结算账户不存在，可能为实时上场新增账户，以账户 昨结算 + 入金 - 出金 作为初始资金
					preBalance = sinkAccount.PreBalance + sinkAccount.Deposit - sinkAccount.Withdraw

					// 新增一条结算记录，以触发后续的boundary更新
					sink.settlements.Store(
						sinkAccount.InvestorID,
						&service.Account{
							Investor: &service.Investor{
								InvestorId: sinkAccount.InvestorID,
							},
							TradingDay: sinkAccount.TradingDay,
							PreBalance: preBalance,
						},
					)
				}

				currBar = sink.newSinkBar()
				currBar.TradingDay = sinkAccount.TradingDay
				currBar.AccountID = sinkAccount.InvestorID
				currBar.Timestamp = sink.barTs
				currBar.Duration = sink.duration.String()
				currBar.Open = preBalance
				currBar.Highest = preBalance
				currBar.Lowest = preBalance
				currBar.Close = preBalance

				sink.barCache[sinkAccount.InvestorID] = currBar
			}

			// bar数据切换
			if sinkAccount.Timestamp.After(currBar.Timestamp) {
				currBar.TradingDay = sinkAccount.TradingDay
				currBar.Timestamp = sink.barTs

				if sink.mode == FirstTick {
					currBar.Open = sinkAccount.Balance
					currBar.Highest = sinkAccount.Balance
					currBar.Lowest = sinkAccount.Balance
					// currBar.Close = sinkAccount.Balance
				} else {
					currBar.Open = currBar.Close
					currBar.Highest = currBar.Close
					currBar.Lowest = currBar.Close
					// currBar.Close = sinkAccount.Balance
				}
			}

			if sinkAccount.Balance > currBar.Highest {
				currBar.Highest = sinkAccount.Balance
			} else if sinkAccount.Balance < currBar.Lowest {
				currBar.Lowest = sinkAccount.Balance
			}

			currBar.Close = sinkAccount.Balance

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

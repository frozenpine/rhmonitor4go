package service

import (
	"time"

	rohon "github.com/frozenpine/rhmonitor4go"
)

var (
	privilegeMap = map[rohon.PrivilegeType]PrivilegeType{
		rohon.RH_MONITOR_ADMINISTRATOR: admin,
		rohon.RH_MONITOR_NOMAL:         user,
	}

	currencyMap = map[string]CurrencyID{
		"USD": USD,
		"CNY": CNY,
	}

	bizTypeMap = map[rohon.BusinessType]BusinessType{
		rohon.RH_TRADE_BZTP_Future: future,
		rohon.RH_TRADE_BZTP_Stock:  stock,
	}

	priceTypeMap = map[rohon.OrderPriceType]OrderPriceType{
		rohon.RH_TRADE_OPT_AnyPrice:                any_price,
		rohon.RH_TRADE_OPT_LimitPrice:              limit_price,
		rohon.RH_TRADE_OPT_BestPrice:               best_price,
		rohon.RH_TRADE_OPT_LastPrice:               last_price,
		rohon.RH_TRADE_OPT_LastPricePlusOneTicks:   last_price_plus1,
		rohon.RH_TRADE_OPT_LastPricePlusTwoTicks:   last_price_plus2,
		rohon.RH_TRADE_OPT_LastPricePlusThreeTicks: last_price_plus3,
		rohon.RH_TRADE_OPT_AskPrice1:               ask_price,
		rohon.RH_TRADE_OPT_AskPrice1PlusOneTicks:   ask_price_plus1,
		rohon.RH_TRADE_OPT_AskPrice1PlusTwoTicks:   ask_price_plus2,
		rohon.RH_TRADE_OPT_AskPrice1PlusThreeTicks: ask_price_plus3,
		rohon.RH_TRADE_OPT_BidPrice1:               bid_price,
		rohon.RH_TRADE_OPT_BidPrice1PlusOneTicks:   bid_price_plus1,
		rohon.RH_TRADE_OPT_BidPrice1PlusTwoTicks:   bid_price_plus2,
		rohon.RH_TRADE_OPT_BidPrice1PlusThreeTicks: bid_price_plus3,
		rohon.RH_TRADE_OPT_FiveLevelPrice:          five_level_price,
		rohon.RH_TRADE_STOPLOSS_MARKET:             stop_loss_market,
		rohon.RH_TRADE_STOPLOSS_LIMIT:              stop_loss_limit,
		rohon.RH_TRADE_GTC_LIMIT:                   gtc_limit,
		rohon.RH_TRADE_STOCK_LEND:                  stock_lend,
		rohon.RH_TRADE_STOCK_FINANCING_BUY:         stock_financing_buy,
		rohon.RH_TRADE_REPAY_STOCK:                 repay_stock_type,
		rohon.RH_TRADE_ETF_PURCHASE:                etf_purchase,
		rohon.RH_TRADE_ETF_REDEMPTION:              etf_redemption,
	}

	directionMap = map[rohon.Direction]Direction{
		rohon.RH_TRADE_D_Buy:  buy,
		rohon.RH_TRADE_D_Sell: sell,
	}

	timeConditionMap = map[rohon.TimeCondition]TimeCondition{
		rohon.RH_TRADE_TC_IOC: immediate_or_cancel,
		rohon.RH_TRADE_TC_GFS: good_for_section,
		rohon.RH_TRADE_TC_GFD: good_for_day,
		rohon.RH_TRADE_TC_GTD: good_till_date,
		rohon.RH_TRADE_TC_GTC: good_till_canceled,
		rohon.RH_TRADE_TC_GFA: good_for_auction,
	}

	volumeConditionMap = map[rohon.VolumeCondition]VolumeCondition{
		rohon.RH_TRADE_VC_AV: any_volume,
		rohon.RH_TRADE_VC_MV: min_volume,
		rohon.RH_TRADE_VC_CV: all_volume,
	}

	contingentConditionMap = map[rohon.ContingentCondition]ContingentCondition{
		rohon.RH_TRADE_CC_Immediately:                    immediately,
		rohon.RH_TRADE_CC_Touch:                          touch,
		rohon.RH_TRADE_CC_TouchProfit:                    touch_profit,
		rohon.RH_TRADE_CC_ParkedOrder:                    parked_order,
		rohon.RH_TRADE_CC_LastPriceGreaterThanStopPrice:  last_price_greate_than_stop_price,
		rohon.RH_TRADE_CC_LastPriceGreaterEqualStopPrice: last_price_greate_equal_stop_price,
		rohon.RH_TRADE_CC_LastPriceLesserThanStopPrice:   last_price_less_than_stop_price,
		rohon.RH_TRADE_CC_LastPriceLesserEqualStopPrice:  last_price_less_equal_stop_price,
		rohon.RH_TRADE_CC_AskPriceGreaterThanStopPrice:   ask_price_greate_than_stop_price,
		rohon.RH_TRADE_CC_AskPriceGreaterEqualStopPrice:  ask_price_greate_equal_stop_price,
		rohon.RH_TRADE_CC_AskPriceLesserThanStopPrice:    ask_price_less_than_stop_price,
		rohon.RH_TRADE_CC_AskPriceLesserEqualStopPrice:   ask_price_less_equal_stop_price,
		rohon.RH_TRADE_CC_BidPriceGreaterThanStopPrice:   bid_price_greate_than_stop_price,
		rohon.RH_TRADE_CC_BidPriceGreaterEqualStopPrice:  bid_price_greate_equal_stop_price,
		rohon.RH_TRADE_CC_BidPriceLesserThanStopPrice:    bid_price_less_than_stop_price,
		rohon.RH_TRADE_CC_BidPriceLesserEqualStopPrice:   bid_price_less_equal_stop_price,
		rohon.RH_TRADE_CC_CloseYDFirst:                   close_yesterday_first,
	}

	forceCloseMap = map[rohon.ForceCloseReason]ForceCloseReason{
		rohon.RH_TRADE_FCC_NotForceClose:           not_force_close,
		rohon.RH_TRADE_FCC_LackDeposit:             lack_deposit,
		rohon.RH_TRADE_FCC_ClientOverPositionLimit: client_over_position_limit,
		rohon.RH_TRADE_FCC_MemberOverPositionLimit: member_over_position_limit,
		rohon.RH_TRADE_FCC_NotMultiple:             not_multiple,
		rohon.RH_TRADE_FCC_Violation:               violation,
		rohon.RH_TRADE_FCC_Other:                   other,
		rohon.RH_TRADE_FCC_PersonDeliv:             person_deliv,
	}

	ordSubmitStatusMap = map[rohon.OrderSubmitStatus]OrderSubmitStatus{
		rohon.RH_TRADE_OSS_InsertSubmitted: insert_submitted,
		rohon.RH_TRADE_OSS_CancelSubmitted: cancel_submitted,
		rohon.RH_TRADE_OSS_ModifySubmitted: modify_submitted,
		rohon.RH_TRADE_OSS_Accepted:        accepted,
		rohon.RH_TRADE_OSS_InsertRejected:  insert_rejected,
		rohon.RH_TRADE_OSS_CancelRejected:  cancel_rejected,
		rohon.RH_TRADE_OSS_ModifyRejected:  modify_rejected,
	}

	ordSrcMap = map[rohon.OrderSource]OrderSource{
		rohon.RH_TRADE_OSRC_Participant:             participant,
		rohon.RH_TRADE_OSRC_Administrator:           administrator,
		rohon.RH_TRADE_OSRC_QryOrder:                query_order,
		rohon.RH_TRADE_OSRC_MonitorForceOrder:       monitor_force_order,
		rohon.RH_TRADE_OSRC_RiskForceOrder:          risk_force_order,
		rohon.RH_TRADE_OSRC_MonitorThirdOrder:       monitor_third_order,
		rohon.RH_TRADE_OSRC_RealObjThirdOrder:       real_obj_third_order,
		rohon.RH_TRADE_OSRC_ServerCondiOrder:        server_condition_order,
		rohon.RH_TRADE_OSRC_ServerLossOrder:         server_loss_order,
		rohon.RH_TRADE_OSRC_ServerProfitOrder:       server_profit_order,
		rohon.RH_TRADE_OSRC_ServerLossEnsureOrder:   server_loss_ensure_order,
		rohon.RH_TRADE_OSRC_ServerProfitEnsureOrder: server_profit_ensure_order,
		rohon.RH_TRADE_OSRC_ServerParkedOrder:       server_parked_order,
	}

	ordStatusMap = map[rohon.OrderStatus]OrderStatus{
		rohon.RH_TRADE_OST_AllTraded:             all_traded,
		rohon.RH_TRADE_OST_PartTradedQueueing:    part_traded_queueing,
		rohon.RH_TRADE_OST_PartTradedNotQueueing: part_traded_not_queueing,
		rohon.RH_TRADE_OST_NoTradeQueueing:       no_trade_queueing,
		rohon.RH_TRADE_OST_NoTradeNotQueueing:    no_trade_not_queueing,
		rohon.RH_TRADE_OST_Canceled:              canceled,
		rohon.RH_TRADE_OST_Unknown:               unknown,
		rohon.RH_TRADE_OST_NotTouched:            not_touched,
		rohon.RH_TRADE_OST_Touched:               touched,
		rohon.RH_TRADE_OST_Submitted:             submitted,
		rohon.RH_TRADE_OST_Amending:              amending,
	}

	ordTypeMap = map[rohon.OrderType]OrderType{
		rohon.RH_TRADE_ORDT_Normal:                normal,
		rohon.RH_TRADE_ORDT_DeriveFromQuote:       derive_from_quote,
		rohon.RH_TRADE_ORDT_DeriveFromCombination: derive_from_combination,
		rohon.RH_TRADE_ORDT_Combination:           combination,
		rohon.RH_TRADE_ORDT_ConditionalOrder:      conditional_order,
		rohon.RH_TRADE_ORDT_Swap:                  swap,
		rohon.RH_TRADE_ORDT_FinancingBuy:          financing_buy,
		rohon.RH_TRADE_ORDT_SellRepayMoney:        sell_repay_money,
		rohon.RH_TRADE_ORDT_FinancingSell:         finacing_sell,
		rohon.RH_TRADE_ORDT_RepayStock:            repay_stock,
	}

	tradingRoleMap = map[rohon.TradingRole]TradingRole{
		rohon.RH_TRADE_ER_Broker: broker,
		rohon.RH_TRADE_ER_Host:   host,
		rohon.RH_TRADE_ER_Maker:  maker,
	}

	offsetFlagMap = map[rohon.OffsetFlag]OffsetFlag{
		rohon.RH_TRADE_OF_Open:            open,
		rohon.RH_TRADE_OF_Close:           close,
		rohon.RH_TRADE_OF_ForceClose:      force_close,
		rohon.RH_TRADE_OF_CloseToday:      close_today,
		rohon.RH_TRADE_OF_CloseYesterday:  close_yesterday,
		rohon.RH_TRADE_OF_ForceOff:        force_off,
		rohon.RH_TRADE_OF_LocalForceClose: local_force_off,
	}

	hedgeFlagMap = map[rohon.HedgeFlag]HedgeFlag{
		rohon.RH_TRADE_HF_Speculation: speculation,
		rohon.RH_TRADE_HF_Arbitrage:   arbitrage,
		rohon.RH_TRADE_HF_Hedge:       hedge,
		rohon.RH_TRADE_HF_MarketMaker: market_maker,
	}

	tradeTypeMap = map[rohon.TradeType]TradeType{
		rohon.RH_TRADE_TRDT_SplitCombination:   split_combination,
		rohon.RH_TRADE_TRDT_Common:             common,
		rohon.RH_TRADE_TRDT_OptionsExecution:   options_execution,
		rohon.RH_TRADE_TRDT_OTC:                otc,
		rohon.RH_TRADE_TRDT_EFPDerived:         efp_derived,
		rohon.RH_TRADE_TRDT_CombinationDerived: combination_derived,
		rohon.RH_TRADE_TRDT_FinancingBuy:       finacing_buy,
		rohon.RH_TRADE_TRDT_RepayStock_Auto:    repay_stock_auto,
		rohon.RH_TRADE_TRDT_RepayStock_Manual:  repay_stock_manual,
	}

	priceSrcMap = map[rohon.PriceSource]PriceSource{
		rohon.RH_TRADE_PSRC_LastPrice: ps_last_price,
		rohon.RH_TRADE_PSRC_Buy:       ps_buy,
		rohon.RH_TRADE_PSRC_Sell:      ps_sell,
	}

	tradeSrcMap = map[rohon.TradeSource]TradeSource{
		rohon.TRH_TSRC_NORMAL: ts_normal,
		rohon.TRH_TSRC_QUERY:  ts_query,
	}
)

func ConvertRspLogin(login *rohon.RspUserLogin) *RspUserLogin {
	return &RspUserLogin{
		UserId:        login.UserID,
		TradingDay:    login.TradingDay,
		LoginTime:     login.LoginTime,
		PrivilegeType: privilegeMap[login.PrivilegeType],
		PrivilegeInfo: login.InfoPrivilegeType.ToDict(),
	}
}

func ConvertRspLogout(logout *rohon.RspUserLogout) *RspUserLogout {
	return &RspUserLogout{
		UserId: logout.UserID,
	}
}

func ConvertInvestor(inv *rohon.Investor) *Investor {
	return &Investor{
		BrokerId:   inv.BrokerID,
		InvestorId: inv.InvestorID,
	}
}

func ConvertOrder(ord *rohon.Order) *Order {
	return &Order{
		Investor: &Investor{
			BrokerId:   ord.BrokerID,
			InvestorId: ord.InvestorID,
		},
		InstrumentId:         ord.InstrumentID,
		OrderRef:             ord.OrderRef,
		PriceType:            priceTypeMap[ord.PriceType],
		Direction:            directionMap[ord.Direction],
		LimitPrice:           ord.LimitPrice,
		VolumeTotalOrigin:    int32(ord.VolumeTotalOriginal),
		TimeCondition:        timeConditionMap[ord.TimeCondition],
		GtdDate:              ord.GTDDate,
		VolumeCondition:      volumeConditionMap[ord.VolumeCondition],
		MinVolume:            int32(ord.MinVolume),
		ContingentCondition:  contingentConditionMap[ord.ContingentCondition],
		StopPrice:            ord.StopPrice,
		ForceCloseReason:     forceCloseMap[ord.ForceCloseReason],
		IsAutoSuspend:        ord.IsAutoSuspend,
		BusinessUnit:         ord.BusinessUnit,
		RequestId:            int64(ord.RequestID),
		OrderLocalId:         ord.OrderLocalID,
		ExchangeId:           ord.ExchangeID,
		ParticipantId:        ord.ParticipantID,
		ClientId:             ord.ClientID,
		ExchangeInstrumentId: ord.ExchangeInstID,
		TraderId:             ord.TraderID,
		InstallId:            int32(ord.InstallID),
		OrderSubmitStatus:    ordSubmitStatusMap[ord.OrderSubmitStatus],
		NotifySequence:       int32(ord.NotifySequence),
		TradingDay:           ord.TradingDay,
		SettlementId:         int32(ord.SettlementID),
		OrderSysId:           ord.OrderSysID,
		OrderSource:          ordSrcMap[ord.OrderSource],
		OrderStatus:          ordStatusMap[ord.OrderStatus],
		OrderType:            ordTypeMap[ord.OrderType],
		VolumeTraded:         int32(ord.VolumeTraded),
		VolumeTotal:          int32(ord.VolumeTotal),
		InsertDate:           ord.InsertDate,
		InsertTime:           ord.InsertTime,
		ActiveTime:           ord.ActiveTime,
		SuspendTime:          ord.SuspendTime,
		UpdateTime:           ord.UpdateTime,
		CancelTime:           ord.CancelTime,
		ActiveTraderId:       ord.ActiveTraderID,
		ClearingPartId:       ord.ClearingPartID,
		SequenceNo:           int32(ord.SequenceNo),
		FrontId:              int32(ord.FrontID),
		SessionId:            int32(ord.SessionID),
		UserProductInfo:      ord.UserProductInfo,
		StatusMessage:        ord.StatusMsg,
		UserForceClose:       ord.UserForceClose,
		ActiveUserId:         ord.ActiveUserID,
		BrokerOrderSequence:  int32(ord.BrokerOrderSeq),
		RelativeOrderSysId:   ord.RelativeOrderSysID,
		ZceTotalTradedVolume: int32(ord.ZCETotalTradedVolume),
		IsSwapOrder:          ord.IsSwapOrder,
		BranchId:             ord.BranchID,
		InvestUnitId:         ord.InvestUnitID,
		CurrencyId:           currencyMap[ord.CurrencyID],
		IpAddress:            ord.IPAddress,
		MacAddress:           ord.MACAddress,
	}
}

func ConvertTrade(td *rohon.Trade) *Trade {
	return &Trade{
		Investor: &Investor{
			BrokerId:   td.BrokerID,
			InvestorId: td.InvestorID,
		},
		InstrumentId:         td.InstrumentID,
		OrderRef:             td.OrderRef,
		UserId:               td.UserID,
		ExchangeId:           td.ExchangeID,
		TradeId:              td.TradeID,
		Direction:            directionMap[td.Direction],
		OrderSysId:           td.OrderSysID,
		ParticipantId:        td.ParticipantID,
		ClientId:             td.ClientID,
		TradingRole:          tradingRoleMap[td.TradingRole],
		ExchangeInstrumentId: td.ExchangeInstID,
		OffsetFlag:           offsetFlagMap[td.OffsetFlag],
		HedgeFlag:            hedgeFlagMap[td.HedgeFlag],
		Price:                td.Price,
		Volume:               int32(td.Volume),
		TradeDate:            td.TradeDate,
		TradeTime:            td.TradeTime,
		TradeType:            tradeTypeMap[td.TradeType],
		PriceSource:          priceSrcMap[td.PriceSource],
		TraderId:             td.TraderID,
		OrderLocalId:         td.OrderLocalID,
		ClearingPartId:       td.ClearingPartID,
		BusinessUnit:         td.BusinessUnit,
		SequenceNo:           int32(td.SequenceNo),
		TradingDay:           td.TradingDay,
		SettlementId:         int32(td.SettlementID),
		BrokerOrderSeqence:   int32(td.BrokerOrderSeq),
		TradeSource:          tradeSrcMap[td.TradeSource],
		InvestorUnitId:       td.InvestUnitID,
	}
}

func ConvertPosition(pos *rohon.Position) *Position {
	return &Position{
		Investor: &Investor{
			BrokerId:   pos.BrokerID,
			InvestorId: pos.InvestorID,
		},
		ProductId:         pos.ProductID,
		InstrumentId:      pos.InstrumentID,
		HedgeFlag:         hedgeFlagMap[pos.HedgeFlag],
		Direction:         directionMap[pos.Direction],
		Volume:            int32(pos.Volume),
		Margin:            pos.Margin,
		AvgOpenPriceByVol: pos.AvgOpenPriceByVol,
		AvgOpenPrice:      pos.AvgOpenPrice,
		TodayVolume:       int32(pos.TodayVolume),
		FrozenVolume:      int32(pos.FrozenVolume),
		EntryType:         uint32(pos.EntryType),
	}
}

func ConvertAccount(acct *rohon.Account) *Account {
	return &Account{
		Investor: &Investor{
			BrokerId:   acct.BrokerID,
			InvestorId: acct.AccountID,
		},
		PreCredit:              acct.PreCredit,
		PreDeposit:             acct.PreDeposit,
		PreBalance:             acct.PreBalance,
		PreMargin:              acct.PreMargin,
		InterestBase:           acct.InterestBase,
		Interest:               acct.Interest,
		Deposit:                acct.Deposit,
		Withdraw:               acct.Withdraw,
		FrozenMargin:           acct.FrozenMargin,
		FrozenCash:             acct.FrozenCash,
		FrozenCommission:       acct.FrozenCommission,
		CurrentMargin:          acct.CurrMargin,
		CashIn:                 acct.CashIn,
		Commission:             acct.Commission,
		CloseProfit:            acct.CloseProfit,
		PositionProfit:         acct.PositionProfit,
		Balance:                acct.Balance,
		Available:              acct.Available,
		WithdrawQuota:          acct.WithdrawQuota,
		Reserve:                acct.Reserve,
		TradingDay:             acct.TradingDay,
		SettlementId:           int32(acct.SettlementID),
		Credit:                 acct.Credit,
		ExchangeMargin:         acct.ExchangeMargin,
		DeliveryMargin:         acct.DeliveryMargin,
		ExchangeDeliveryMargin: acct.ExchangeDeliveryMargin,
		ReserveBalance:         acct.ReserveBalance,
		CurrencyId:             currencyMap[acct.CurrencyID],
		MortgageInfo: &FundMortgage{
			PreIn:       acct.PreFundMortgageIn,
			PreOut:      acct.PreFundMortgageOut,
			PreMortgage: acct.PreMortgage,
			CurrentIn:   acct.FundMortgageIn,
			CurrentOut:  acct.FundMortgageOut,
			Mortgage:    acct.Mortgage,
			Available:   acct.FundMortgageAvailable,
			Mortgagable: acct.MortgageableFund,
		},
		SpecProductInfo: &SpecProduct{
			Margin:              acct.SpecProductMargin,
			FrozenMargin:        acct.SpecProductFrozenMargin,
			Commission:          acct.SpecProductCommission,
			FrozenCommission:    acct.SpecProductFrozenCommission,
			PositionProfit:      acct.SpecProductPositionProfit,
			CloseProfit:         acct.SpecProductCloseProfit,
			PositionProfitByAlg: acct.SpecProductPositionProfitByAlg,
			ExchangeMargin:      acct.SpecProductExchangeMargin,
		},
		BusinessType:      bizTypeMap[acct.BizType],
		FrozenSwap:        acct.FrozenSwap,
		RemainSwap:        acct.RemainSwap,
		StockMarketValue:  acct.TotalStockMarketValue,
		OptionMarketValue: acct.TotalOptionMarketValue,
		DynamicMoney:      acct.DynamicMoney,
		Premium:           acct.Premium,
		MarketValueEquity: acct.MarketValueEquity,
		Timestamp:         time.Now().UnixMicro(),
	}
}

package rohon

type Reason uint16

//go:generate stringer -type Reason -linecomment
const (
	NetReadFailed  Reason = 0x1001 // 网络读失败
	NetWriteFailed Reason = 0x1002 // 网络写失败
	HBTimeout      Reason = 0x2001 // 接收心跳超时
	HBSendFaild    Reason = 0x2002 // 发送心跳失败
	InvalidPacket  Reason = 0x2003 // 收到错误报文
)

type PrivilegeType uint8

const (
	RH_MONITOR_ADMINISTRATOR PrivilegeType = '0' + iota
	RH_MONITOR_NOMAL
)

type BusinessType uint8

//go:generate stringer -type BusinessType -linecomment
const (
	RH_TRADE_BZTP_Future BusinessType = '1' + iota // 期货
	RH_TRADE_BZTP_Stock                            // 证券
)

type OffsetFlag uint8

//go:generate stringer -type OffsetFlag -linecomment
const (
	RH_TRADE_OF_Open            OffsetFlag = '0' + iota // 开仓
	RH_TRADE_OF_Close                                   // 平仓
	RH_TRADE_OF_ForceClose                              // 强平
	RH_TRADE_OF_CloseToday                              // 平今
	RH_TRADE_OF_CloseYesterday                          // 平昨
	RH_TRADE_OF_ForceOff                                // 强减
	RH_TRADE_OF_LocalForceClose                         // 本地强减
)

type HedgeFlag uint8

//go:generate stringer -type HedgeFlag -linecomment
const (
	RH_TRADE_HF_Speculation HedgeFlag = '1' + iota // 投机
	RH_TRADE_HF_Arbitrage                          // 套利
	RH_TRADE_HF_Hedge                              // 套保
	_
	RH_TRADE_HF_MarketMaker // 做市商
)

type Direction uint8

//go:generate stringer -type Direction -linecomment
const (
	RH_TRADE_D_Buy  Direction = '0' + iota // 买
	RH_TRADE_D_Sell                        // 卖
)

type OrderPriceType uint8

//go:generate stringer -type OrderPriceType -linecomment
const (
	RH_TRADE_OPT_AnyPrice                OrderPriceType = '1' + iota // 任意价
	RH_TRADE_OPT_LimitPrice                                          // 限价
	RH_TRADE_OPT_BestPrice                                           // 最优价
	RH_TRADE_OPT_LastPrice                                           // 最新价
	RH_TRADE_OPT_LastPricePlusOneTicks                               // 最新价浮动上浮1个ticks
	RH_TRADE_OPT_LastPricePlusTwoTicks                               // 最新价浮动上浮2个ticks
	RH_TRADE_OPT_LastPricePlusThreeTicks                             // 最新价浮动上浮3个ticks
	RH_TRADE_OPT_AskPrice1                                           // 卖一价
	RH_TRADE_OPT_AskPrice1PlusOneTicks                               // 卖一价浮动上浮1个ticks
	RH_TRADE_OPT_AskPrice1PlusTwoTicks   OrderPriceType = 'A' + iota // 卖一价浮动上浮2个ticks
	RH_TRADE_OPT_AskPrice1PlusThreeTicks                             // 卖一价浮动上浮3个ticks
	RH_TRADE_OPT_BidPrice1                                           // 买一价
	RH_TRADE_OPT_BidPrice1PlusOneTicks                               // 买一价浮动上浮1个ticks
	RH_TRADE_OPT_BidPrice1PlusTwoTicks                               // 买一价浮动上浮2个ticks
	RH_TRADE_OPT_BidPrice1PlusThreeTicks                             // 买一价浮动上浮3个ticks
	RH_TRADE_OPT_FiveLevelPrice                                      // 五档价
	RH_TRADE_STOPLOSS_MARKET                                         // 止损市价
	RH_TRADE_STOPLOSS_LIMIT                                          // 止损限价
	RH_TRADE_GTC_LIMIT                                               // 长效单
	RH_TRADE_STOCK_LEND                                              // 一键锁券
	RH_TRADE_STOCK_FINANCING_BUY                                     // 融资买入单
	RH_TRADE_REPAY_STOCK                                             // 现券还券单
	RH_TRADE_ETF_PURCHASE                                            // ETF申购
	RH_TRADE_ETF_REDEMPTION                                          // ETF赎回
)

type TimeCondition uint8

//go:generate stringer -type TimeCondition -linecomment
const (
	RH_TRADE_TC_IOC TimeCondition = '1' + iota // 立即完成
	RH_TRADE_TC_GFS                            // 本节有效
	RH_TRADE_TC_GFD                            // 当日有效
	RH_TRADE_TC_GTD                            // 指定日期前有效
	RH_TRADE_TC_GTC                            // 撤销前有效
	RH_TRADE_TC_GFA                            // 集合竞价有效
)

type VolumeCondition uint8

//go:generate stringer -type VolumeCondition -linecomment
const (
	RH_TRADE_VC_AV VolumeCondition = '1' + iota // 任何数量
	RH_TRADE_VC_MV                              // 最小数量
	RH_TRADE_VC_CV                              // 全部数量
)

type ContingentCondition uint8

//go:generate stringer -type ContingentCondition -linecomment
const (
	RH_TRADE_CC_Immediately                    ContingentCondition = '1' + iota // 立即
	RH_TRADE_CC_Touch                                                           // 止损
	RH_TRADE_CC_TouchProfit                                                     // 止赢
	RH_TRADE_CC_ParkedOrder                                                     // 预埋单
	RH_TRADE_CC_LastPriceGreaterThanStopPrice                                   // 最新价大于条件价
	RH_TRADE_CC_LastPriceGreaterEqualStopPrice                                  // 最新价大于等于条件价
	RH_TRADE_CC_LastPriceLesserThanStopPrice                                    // 最新价小于条件价
	RH_TRADE_CC_LastPriceLesserEqualStopPrice                                   // 最新价小于等于条件价
	RH_TRADE_CC_AskPriceGreaterThanStopPrice                                    // 卖一价大于条件价
	RH_TRADE_CC_AskPriceGreaterEqualStopPrice  ContingentCondition = 'A' + iota // 卖一价大于等于条件价
	RH_TRADE_CC_AskPriceLesserThanStopPrice                                     // 卖一价小于条件价
	RH_TRADE_CC_AskPriceLesserEqualStopPrice                                    // 卖一价小于等于条件价
	RH_TRADE_CC_BidPriceGreaterThanStopPrice                                    // 买一价大于条件价
	RH_TRADE_CC_BidPriceGreaterEqualStopPrice                                   // 买一价大于等于条件价
	RH_TRADE_CC_BidPriceLesserThanStopPrice                                     // 买一价小于条件价
	RH_TRADE_CC_BidPriceLesserEqualStopPrice   ContingentCondition = 'H'        // 买一价小于等于条件价
	RH_TRADE_CC_CloseYDFirst                   ContingentCondition = 'Z'        // 开空指令转换为平昨仓优先
)

type ForceCloseReason uint8

//go:generate stringer -type ForceCloseReason -linecomment
const (
	RH_TRADE_FCC_NotForceClose           ForceCloseReason = '0' + iota // 非强平
	RH_TRADE_FCC_LackDeposit                                           // 资金不足
	RH_TRADE_FCC_ClientOverPositionLimit                               // 客户超仓
	RH_TRADE_FCC_MemberOverPositionLimit                               // 会员超仓
	RH_TRADE_FCC_NotMultiple                                           // 持仓非整数倍
	RH_TRADE_FCC_Violation                                             // 违规
	RH_TRADE_FCC_Other                                                 // 其它
	RH_TRADE_FCC_PersonDeliv                                           // 自然人临近交割
)

type OrderSubmitStatus uint8

//go:generate stringer -type OrderSubmitStatus -linecomment
const (
	RH_TRADE_OSS_InsertSubmitted OrderSubmitStatus = '0' + iota // 已经提交
	RH_TRADE_OSS_CancelSubmitted                                // 撤单已经提交
	RH_TRADE_OSS_ModifySubmitted                                // 修改已经提交
	RH_TRADE_OSS_Accepted                                       // 已经接受
	RH_TRADE_OSS_InsertRejected                                 // 报单已拒绝
	RH_TRADE_OSS_CancelRejected                                 // 撤单已拒绝
	RH_TRADE_OSS_ModifyRejected                                 // 改单已拒绝
)

type OrderSource uint8

//go:generate stringer -type OrderSource -linecomment
const (
	RH_TRADE_OSRC_Participant             OrderSource = '0' + iota // 来自参与者
	RH_TRADE_OSRC_Administrator                                    // 来自管理员
	RH_TRADE_OSRC_QryOrder                                         // 查询报单
	RH_TRADE_OSRC_MonitorForceOrder                                // 来自于风控的强平单
	RH_TRADE_OSRC_RiskForceOrder                                   // 触发风险后的强平单
	RH_TRADE_OSRC_MonitorThirdOrder                                // 风控端第三方报单
	RH_TRADE_OSRC_RealObjThirdOrder                                // 资金账户外部报单后自动映射的报单
	RH_TRADE_OSRC_ServerCondiOrder                                 // 服务器条件单
	RH_TRADE_OSRC_ServerLossOrder                                  // 服务器止损单
	RH_TRADE_OSRC_ServerProfitOrder                                // 服务器止盈单
	RH_TRADE_OSRC_ServerLossEnsureOrder   OrderSource = 'a' + iota // 服务器止损追单
	RH_TRADE_OSRC_ServerProfitEnsureOrder                          // 服务器止盈追单
	RH_TRADE_OSRC_ServerParkedOrder                                // 服务器预埋单
)

type OrderStatus uint8

//go:generate stringer -type OrderStatus -linecomment
const (
	RH_TRADE_OST_AllTraded             OrderStatus = '0' + iota // 全部成交
	RH_TRADE_OST_PartTradedQueueing                             // 部分成交还在队列中
	RH_TRADE_OST_PartTradedNotQueueing                          // 部分成交不在队列中
	RH_TRADE_OST_NoTradeQueueing                                // 未成交还在队列中
	RH_TRADE_OST_NoTradeNotQueueing                             // 未成交不在队列中
	RH_TRADE_OST_Canceled                                       // 撤单
	RH_TRADE_OST_Unknown               OrderStatus = 'a' + iota // 未知
	RH_TRADE_OST_NotTouched                                     // 尚未触发
	RH_TRADE_OST_Touched                                        // 已触发
	RH_TRADE_OST_Submitted                                      // 已提交
	RH_TRADE_OST_Amending              OrderStatus = 'm'        // 正在修改
)

type OrderType uint8

//go:generate stringer -type OrderType -linecomment
const (
	RH_TRADE_ORDT_Normal                OrderType = '0' + iota // 正常
	RH_TRADE_ORDT_DeriveFromQuote                              // 报价衍生
	RH_TRADE_ORDT_DeriveFromCombination                        // 组合衍生
	RH_TRADE_ORDT_Combination                                  // 组合报单
	RH_TRADE_ORDT_ConditionalOrder                             // 条件单
	RH_TRADE_ORDT_Swap                                         // 互换单
	RH_TRADE_ORDT_FinancingBuy          OrderType = 'A' + iota // 融资买入单
	RH_TRADE_ORDT_SellRepayMoney                               // 卖券还款
	RH_TRADE_ORDT_FinancingSell                                // 融资平仓
	RH_TRADE_ORDT_RepayStock            OrderType = 'R'        // 现券还券
)

type TradingRole uint8

//go:generate stringer -type TradingRole -linecomment
const (
	RH_TRADE_ER_Broker TradingRole = '1' + iota // 代理
	RH_TRADE_ER_Host                            // 自营
	RH_TRADE_ER_Maker                           // 做市商
)

type TradeType uint8

//go:generate stringer -type TradeType -linecomment
const (
	RH_TRADE_TRDT_SplitCombination   TradeType = '#'        // 组合持仓拆分为单一持仓
	RH_TRADE_TRDT_Common             TradeType = '0' + iota // 普通成交
	RH_TRADE_TRDT_OptionsExecution                          // 期权执行
	RH_TRADE_TRDT_OTC                                       // OTC成交
	RH_TRADE_TRDT_EFPDerived                                // 期转现衍生成交
	RH_TRADE_TRDT_CombinationDerived                        // 组合衍生成交
	RH_TRADE_TRDT_FinancingBuy       TradeType = 'F'        // 融资买入成交
	RH_TRADE_TRDT_RepayStock_Auto    TradeType = 'R'        // 卖平今的现券还券
	RH_TRADE_TRDT_RepayStock_Manual  TradeType = 'S'        // 正常的现券还券指令
)

type PriceSource uint8

//go:generate stringer -type PriceSource -linecomment
const (
	RH_TRADE_PSRC_LastPrice PriceSource = '0' + iota // 前成交价
	RH_TRADE_PSRC_Buy                                // 买委托价
	RH_TRADE_PSRC_Sell                               // 卖委托价
)

type TradeSource uint8

//go:generate stringer -type TradeSource -linecomment
const (
	TRH_TSRC_NORMAL TradeSource = '0' + iota // 来自交易所回报
	TRH_TSRC_QUERY                           // 来自查询
)

type AccountType uint8

//go:generate stringer -type AccountType -linecomment
const (
	RH_ACCOUNTTYPE_VIRTUAL   AccountType = '0' + iota // 虚拟账户
	RH_ACCOUNTTYPE_REAL                               // 真实账户
	RH_ACCOUNTTYPE_REALGROUP                          // 资金账户组
)

type SubInfoType int

//go:generate stringer -type SubInfoType -linecomment
const (
	RHMonitorSubPushInfoType_Order SubInfoType = 1 << iota // 委托回报
	RHMonitorSubPushInfoType_Trade                         // 成交回报
)

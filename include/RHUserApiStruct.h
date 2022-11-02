/////////////////////////////////////////////////////////////////////////
///@system 融航期货交易平台
///@company 上海融航信息技术有限公司
///@file RHUserApiStruct.h
///@brief 定义了客户端接口使用的业务数据结构
///20180910 create by Haosc
/////////////////////////////////////////////////////////////////////////

#if !defined(RH_TRADESTRUCT_H)
#define RH_TRADESTRUCT_H

#if _MSC_VER > 1000
#pragma once
#endif // _MSC_VER > 1000

#include "RHUserApiDataType.h"

///用户登录请求
struct CRHReqUserLoginField
{
	///交易日
	TRHDateType	TradingDay;
	///经纪公司代码
	TRHBrokerIDType	BrokerID;
	///用户代码
	TRHUserIDType	UserID;
	///密码
	TRHPasswordType	Password;
	///用户端产品信息
	TRHProductInfoType	UserProductInfo;
	///接口端产品信息
	TRHProductInfoType	InterfaceProductInfo;
	///协议信息
	TRHProtocolInfoType	ProtocolInfo;
	///Mac地址
	TRHMacAddressType	MacAddress;
	///动态密码
	TRHPasswordType	OneTimePassword;
	///终端IP地址
	TRHIPAddressType	ClientIPAddress;
	///扩展实例ID，BrokerID和UserID不足以区分多个实例
	TRHTIDType		InstanceID;
	///登录备注
	TRHLoginRemarkType	LoginRemark;
};

///用户登录应答
struct CRHRspUserLoginField
{
	///交易日
	TRHDateType	TradingDay;
	///登录成功时间
	TRHTimeType	LoginTime;
	///经纪公司代码
	TRHBrokerIDType	BrokerID;
	///用户代码
	TRHUserIDType	UserID;
	///交易系统名称
	TRHSystemNameType	SystemName;
	///前置编号
	TRHFrontIDType	FrontID;
	///会话编号
	TRHSessionIDType	SessionID;
	///最大报单引用
	TRHOrderRefType	MaxOrderRef;
	///上期所时间
	TRHTimeType	SHFETime;
	///大商所时间
	TRHTimeType	DCETime;
	///郑商所时间
	TRHTimeType	CZCETime;
	///中金所时间
	TRHTimeType	FFEXTime;
	///能源中心时间
	TRHTimeType	INETime;
	///扩展实例ID，BrokerID和UserID不足以区分多个实例
	TRHTIDType		InstanceID;
};

///用户登出请求
struct CRHUserLogoutField
{
	///经纪公司代码
	TRHBrokerIDType	BrokerID;
	///用户代码
	TRHUserIDType	UserID;
};

///强制交易员退出
struct CRHForceUserLogoutField
{
	///经纪公司代码
	TRHBrokerIDType	BrokerID;
	///用户代码
	TRHUserIDType	UserID;
};

struct CRHMonitorReqUserLoginField 
{
	//风控账号
	TRHUserIDType		  UserID;
	//风控密码
	TRHPasswordType		  Password;
	//MAC地址
	TRHMacAddressType	  MacAddress;
};

struct CRHMonitorUserLogoutField 
{
	//风控账号
	TRHUserIDType		  UserID;
};


struct CRHMonitorRspUserLoginField
{
	//风控账号
	TRHUserIDType		  UserID;
	//账户权限
	TRHPrivilegeType		  PrivilegeType;
	//信息查看权限
	TRHInfoPrivilegeType	  InfoPrivilegeType;
	///交易日
	TRHDateType			  TradingDay;
	///登录成功时间
	TRHTimeType			  LoginTime;
};

struct CRHMonitorQryMonitorUser
{
	//风控账号
	TRHUserIDType		  UserID;
};

struct CRHMonitorRspMonitorUser 
{
	//投资者
	TRHInvestorIDType	  InvestorID;
	//经纪公司代码
	TRHBrokerIDType		  BrokerID;
};

///响应信息
struct CRHRspInfoField
{
	///错误代码
	TRHErrorIDType	ErrorID;
	///错误信息
	TRHErrorMsgType	ErrorMsg;
};

///资金账户
struct CRHTradingAccountField
{
	///经纪公司代码
	TRHBrokerIDType	BrokerID;
	///投资者帐号
	TRHAccountIDType	AccountID;
	///上次质押金额
	TRHMoneyType	PreMortgage;
	///上次信用额度
	TRHMoneyType	PreCredit;
	///上次存款额
	TRHMoneyType	PreDeposit;
	///上次结算准备金
	TRHMoneyType	PreBalance;
	///上次占用的保证金
	TRHMoneyType	PreMargin;
	///利息基数
	TRHMoneyType	InterestBase;
	///利息收入
	TRHMoneyType	Interest;
	///入金金额
	TRHMoneyType	Deposit;
	///出金金额
	TRHMoneyType	Withdraw;
	///冻结的保证金
	TRHMoneyType	FrozenMargin;
	///冻结的资金
	TRHMoneyType	FrozenCash;
	///冻结的手续费
	TRHMoneyType	FrozenCommission;
	///当前保证金总额
	TRHMoneyType	CurrMargin;
	///资金差额
	TRHMoneyType	CashIn;
	///手续费
	TRHMoneyType	Commission;
	///平仓盈亏
	TRHMoneyType	CloseProfit;
	///持仓盈亏
	TRHMoneyType	PositionProfit;
	///期货结算准备金
	TRHMoneyType	Balance;
	///可用资金
	TRHMoneyType	Available;
	///可取资金
	TRHMoneyType	WithdrawQuota;
	///基本准备金
	TRHMoneyType	Reserve;
	///交易日
	TRHDateType	TradingDay;
	///结算编号
	TRHSettlementIDType	SettlementID;
	///信用额度
	TRHMoneyType	Credit;
	///质押金额
	TRHMoneyType	Mortgage;
	///交易所保证金
	TRHMoneyType	ExchangeMargin;
	///投资者交割保证金
	TRHMoneyType	DeliveryMargin;
	///交易所交割保证金
	TRHMoneyType	ExchangeDeliveryMargin;
	///保底期货结算准备金
	TRHMoneyType	ReserveBalance;
	///币种代码
	TRHCurrencyIDType	CurrencyID;
	///上次货币质入金额
	TRHMoneyType	PreFundMortgageIn;
	///上次货币质出金额
	TRHMoneyType	PreFundMortgageOut;
	///货币质入金额
	TRHMoneyType	FundMortgageIn;
	///货币质出金额
	TRHMoneyType	FundMortgageOut;
	///货币质押余额
	TRHMoneyType	FundMortgageAvailable;
	///可质押货币金额
	TRHMoneyType	MortgageableFund;
	///特殊产品占用保证金
	TRHMoneyType	SpecProductMargin;
	///特殊产品冻结保证金
	TRHMoneyType	SpecProductFrozenMargin;
	///特殊产品手续费
	TRHMoneyType	SpecProductCommission;
	///特殊产品冻结手续费
	TRHMoneyType	SpecProductFrozenCommission;
	///特殊产品持仓盈亏
	TRHMoneyType	SpecProductPositionProfit;
	///特殊产品平仓盈亏
	TRHMoneyType	SpecProductCloseProfit;
	///根据持仓盈亏算法计算的特殊产品持仓盈亏
	TRHMoneyType	SpecProductPositionProfitByAlg;
	///特殊产品交易所保证金
	TRHMoneyType	SpecProductExchangeMargin;
	///业务类型
	TRHBizTypeType	BizType;
	///延时换汇冻结金额
	TRHMoneyType	FrozenSwap;
	///剩余换汇额度
	TRHMoneyType	RemainSwap;
	//证券持仓市值
	TRHMoneyType TotalStockMarketValue;
	//期权持仓市值
	TRHMoneyType TotalOptionMarketValue;
	//动态权益
	TRHMoneyType DynamicMoney;
	///权利金收支
	TRHMoneyType Premium;
	///市值权益
	TRHMoneyType MarketValueEquity;
};

//持仓监控信息
struct CRHMonitorPositionField
{
	///投资者代码
	TRHInvestorIDType	InvestorID;
	///经纪公司代码
	TRHBrokerIDType		BrokerID;
	///合约类别
	TRHInstrumentIDType	ProductID;
	///合约代码
	TRHInstrumentIDType	InstrumentID;
	///投机套保标志
	TRHHedgeFlagType		HedgeFlag;
	///持仓方向
	TRHDirectionType		Direction;
	///持仓数量
	TRHVolumeType		Volume;
	///持仓保证金
	TRHMoneyType			Margin;
	///逐笔开仓均价
	TRHMoneyType			AvgOpenPriceByVol;
	///逐日开仓均价
	TRHMoneyType			AvgOpenPrice;
	///今仓数量
	TRHVolumeType		TodayVolume;
	///冻结持仓数量
	TRHVolumeType		FrozenVolume;
	///信息类型
	TRHPositionEntryType	EntryType;			
	///昨仓，冻结持仓数量，逐笔持盈，逐笔开仓均价
};

///查询投资者
struct CRHQryInvestorField
{
	///经纪公司代码
	TRHBrokerIDType	BrokerID;
	///投资者代码
	TRHInvestorIDType	InvestorID;
};

struct CRHMonitorQryInvestorPositionField
{
	///投资者代码
	TRHInvestorIDType	InvestorID;
	///账户类别
	//TRHAccountType		AccountType;
	///经纪公司代码
	TRHBrokerIDType		BrokerID;	
	///合约代码
	TRHInstrumentIDType	InstrumentID;
};

struct CRHMonitorQryInvestorMoneyField
{
	///投资者代码
	TRHInvestorIDType	InvestorID;
	///账户类别
	//TRHAccountType		AccountType;
	///经纪公司代码
	TRHBrokerIDType		BrokerID;	
	
};

//风控端强制平仓字段
struct CRHMonitorOffsetOrderField
{
	//投资者
	TRHInvestorIDType	  InvestorID;
	//经纪公司代码
	TRHBrokerIDType		  BrokerID;
	//合约ID
	TRHInstrumentIDType	  InstrumentID;
	//方向
	TRHDirectionType		  Direction;
	//手数
	TRHVolumeType		  volume;
	//价格
	TRHPriceType			  Price;
	///组合开平标志
	TRHCombOffsetFlagType  CombOffsetFlag;
	///组合投机套保标志
	TRHCombHedgeFlagType	  CombHedgeFlag;
};

///报单
struct CRHOrderField
{
	///经纪公司代码
	TRHBrokerIDType	BrokerID;
	///投资者代码
	TRHInvestorIDType	InvestorID;
	///合约代码
	TRHInstrumentIDType	InstrumentID;
	///报单引用
	TRHOrderRefType	OrderRef;
	///用户代码
	TRHUserIDType	UserID;
	///报单价格条件
	TRHOrderPriceTypeType	OrderPriceType;
	///买卖方向
	TRHDirectionType	Direction;
	///组合开平标志
	TRHCombOffsetFlagType	CombOffsetFlag;
	///组合投机套保标志
	TRHCombHedgeFlagType	CombHedgeFlag;
	///价格
	TRHPriceType	LimitPrice;
	///数量
	TRHVolumeType	VolumeTotalOriginal;
	///有效期类型
	TRHTimeConditionType	TimeCondition;
	///GTD日期
	TRHDateType	GTDDate;
	///成交量类型
	TRHVolumeConditionType	VolumeCondition;
	///最小成交量
	TRHVolumeType	MinVolume;
	///触发条件
	TRHContingentConditionType	ContingentCondition;
	///止损价
	TRHPriceType	StopPrice;
	///强平原因
	TRHForceCloseReasonType	ForceCloseReason;
	///自动挂起标志
	TRHBoolType	IsAutoSuspend;
	///业务单元
	TRHBusinessUnitType	BusinessUnit;
	///请求编号
	TRHRequestIDType	RequestID;
	///本地报单编号
	TRHOrderLocalIDType	OrderLocalID;
	///交易所代码
	TRHExchangeIDType	ExchangeID;
	///会员代码
	TRHParticipantIDType	ParticipantID;
	///客户代码
	TRHClientIDType	ClientID;
	///合约在交易所的代码
	TRHExchangeInstIDType	ExchangeInstID;
	///交易所交易员代码
	TRHTraderIDType	TraderID;
	///安装编号
	TRHInstallIDType	InstallID;
	///报单提交状态
	TRHOrderSubmitStatusType	OrderSubmitStatus;
	///报单提示序号
	TRHSequenceNoType	NotifySequence;
	///交易日
	TRHDateType	TradingDay;
	///结算编号
	TRHSettlementIDType	SettlementID;
	///报单编号
	TRHOrderSysIDType	OrderSysID;
	///报单来源
	TRHOrderSourceType	OrderSource;
	///报单状态
	TRHOrderStatusType	OrderStatus;
	///报单类型
	TRHOrderTypeType	OrderType;
	///今成交数量
	TRHVolumeType	VolumeTraded;
	///剩余数量
	TRHVolumeType	VolumeTotal;
	///报单日期
	TRHDateType	InsertDate;
	///委托时间
	TRHTimeType	InsertTime;
	///激活时间
	TRHTimeType	ActiveTime;
	///挂起时间
	TRHTimeType	SuspendTime;
	///最后修改时间
	TRHTimeType	UpdateTime;
	///撤销时间
	TRHTimeType	CancelTime;
	///最后修改交易所交易员代码
	TRHTraderIDType	ActiveTraderID;
	///结算会员编号
	TRHParticipantIDType	ClearingPartID;
	///序号
	TRHSequenceNoType	SequenceNo;
	///前置编号
	TRHFrontIDType	FrontID;
	///会话编号
	TRHSessionIDType	SessionID;
	///用户端产品信息
	TRHProductInfoType	UserProductInfo;
	///状态信息
	TRHErrorMsgType	StatusMsg;
	///用户强评标志
	TRHBoolType	UserForceClose;
	///操作用户代码
	TRHUserIDType	ActiveUserID;
	///经纪公司报单编号
	TRHSequenceNoType	BrokerOrderSeq;
	///相关报单
	TRHOrderSysIDType	RelativeOrderSysID;
	///郑商所成交数量
	TRHVolumeType	ZCETotalTradedVolume;
	///互换单标志
	TRHBoolType	IsSwapOrder;
	///营业部编号
	TRHBranchIDType	BranchID;
	///投资单元代码
	TRHInvestUnitIDType	InvestUnitID;
	///资金账号
	TRHAccountIDType	AccountID;
	///币种代码
	TRHCurrencyIDType	CurrencyID;
	///IP地址
	TRHIPAddressType	IPAddress;
	///Mac地址
	TRHMacAddressType	MacAddress;
};

///成交
struct CRHTradeField
{
	///经纪公司代码
	TRHBrokerIDType	BrokerID;
	///投资者代码
	TRHInvestorIDType	InvestorID;
	///合约代码
	TRHInstrumentIDType	InstrumentID;
	///报单引用
	TRHOrderRefType	OrderRef;
	///用户代码
	TRHUserIDType	UserID;
	///交易所代码
	TRHExchangeIDType	ExchangeID;
	///成交编号
	TRHTradeIDType	TradeID;
	///买卖方向
	TRHDirectionType	Direction;
	///报单编号
	TRHOrderSysIDType	OrderSysID;
	///会员代码
	TRHParticipantIDType	ParticipantID;
	///客户代码
	TRHClientIDType	ClientID;
	///交易角色
	TRHTradingRoleType	TradingRole;
	///合约在交易所的代码
	TRHExchangeInstIDType	ExchangeInstID;
	///开平标志
	TRHOffsetFlagType	OffsetFlag;
	///投机套保标志
	TRHHedgeFlagType	HedgeFlag;
	///价格
	TRHPriceType	Price;
	///数量
	TRHVolumeType	Volume;
	///成交时期
	TRHDateType	TradeDate;
	///成交时间
	TRHTimeType	TradeTime;
	///成交类型
	TRHTradeTypeType	TradeType;
	///成交价来源
	TRHPriceSourceType	PriceSource;
	///交易所交易员代码
	TRHTraderIDType	TraderID;
	///本地报单编号
	TRHOrderLocalIDType	OrderLocalID;
	///结算会员编号
	TRHParticipantIDType	ClearingPartID;
	///业务单元
	TRHBusinessUnitType	BusinessUnit;
	///序号
	TRHSequenceNoType	SequenceNo;
	///交易日
	TRHDateType	TradingDay;
	///结算编号
	TRHSettlementIDType	SettlementID;
	///经纪公司报单编号
	TRHSequenceNoType	BrokerOrderSeq;
	///成交来源
	TRHTradeSourceType	TradeSource;
	///投资单元代码
	TRHInvestUnitIDType	InvestUnitID;
};

//订阅推送信息
struct CRHMonitorSubPushInfo
{
	///投资者代码
	TRHInvestorIDType	InvestorID;
	///账户类别
	TRHAccountType		AccountType;
	///经纪公司代码
	TRHBrokerIDType		BrokerID;
	///订阅类型
	RHMonitorSubPushInfoType SubInfoType;
};

#endif

// Code generated by "stringer -type TradeType -linecomment"; DO NOT EDIT.

package rhmonitor4go

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[RH_TRADE_TRDT_SplitCombination-35]
	_ = x[RH_TRADE_TRDT_Common-49]
	_ = x[RH_TRADE_TRDT_OptionsExecution-50]
	_ = x[RH_TRADE_TRDT_OTC-51]
	_ = x[RH_TRADE_TRDT_EFPDerived-52]
	_ = x[RH_TRADE_TRDT_CombinationDerived-53]
	_ = x[RH_TRADE_TRDT_FinancingBuy-70]
	_ = x[RH_TRADE_TRDT_RepayStock_Auto-82]
	_ = x[RH_TRADE_TRDT_RepayStock_Manual-83]
}

const (
	_TradeType_name_0 = "组合持仓拆分为单一持仓"
	_TradeType_name_1 = "普通成交期权执行OTC成交期转现衍生成交组合衍生成交"
	_TradeType_name_2 = "融资买入成交"
	_TradeType_name_3 = "卖平今的现券还券正常的现券还券指令"
)

var (
	_TradeType_index_1 = [...]uint8{0, 12, 24, 33, 54, 72}
	_TradeType_index_3 = [...]uint8{0, 24, 51}
)

func (i TradeType) String() string {
	switch {
	case i == 35:
		return _TradeType_name_0
	case 49 <= i && i <= 53:
		i -= 49
		return _TradeType_name_1[_TradeType_index_1[i]:_TradeType_index_1[i+1]]
	case i == 70:
		return _TradeType_name_2
	case 82 <= i && i <= 83:
		i -= 82
		return _TradeType_name_3[_TradeType_index_3[i]:_TradeType_index_3[i+1]]
	default:
		return "TradeType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}

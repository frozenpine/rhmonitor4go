// Code generated by "stringer -type ContingentCondition -linecomment"; DO NOT EDIT.

package rhmonitor4go

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[RH_TRADE_CC_Immediately-49]
	_ = x[RH_TRADE_CC_Touch-50]
	_ = x[RH_TRADE_CC_TouchProfit-51]
	_ = x[RH_TRADE_CC_ParkedOrder-52]
	_ = x[RH_TRADE_CC_LastPriceGreaterThanStopPrice-53]
	_ = x[RH_TRADE_CC_LastPriceGreaterEqualStopPrice-54]
	_ = x[RH_TRADE_CC_LastPriceLesserThanStopPrice-55]
	_ = x[RH_TRADE_CC_LastPriceLesserEqualStopPrice-56]
	_ = x[RH_TRADE_CC_AskPriceGreaterThanStopPrice-57]
	_ = x[RH_TRADE_CC_AskPriceGreaterEqualStopPrice-74]
	_ = x[RH_TRADE_CC_AskPriceLesserThanStopPrice-75]
	_ = x[RH_TRADE_CC_AskPriceLesserEqualStopPrice-76]
	_ = x[RH_TRADE_CC_BidPriceGreaterThanStopPrice-77]
	_ = x[RH_TRADE_CC_BidPriceGreaterEqualStopPrice-78]
	_ = x[RH_TRADE_CC_BidPriceLesserThanStopPrice-79]
	_ = x[RH_TRADE_CC_BidPriceLesserEqualStopPrice-72]
	_ = x[RH_TRADE_CC_CloseYDFirst-90]
}

const (
	_ContingentCondition_name_0 = "立即止损止赢预埋单最新价大于条件价最新价大于等于条件价最新价小于条件价最新价小于等于条件价卖一价大于条件价"
	_ContingentCondition_name_1 = "买一价小于等于条件价"
	_ContingentCondition_name_2 = "卖一价大于等于条件价卖一价小于条件价卖一价小于等于条件价买一价大于条件价买一价大于等于条件价买一价小于条件价"
	_ContingentCondition_name_3 = "开空指令转换为平昨仓优先"
)

var (
	_ContingentCondition_index_0 = [...]uint8{0, 6, 12, 18, 27, 51, 81, 105, 135, 159}
	_ContingentCondition_index_2 = [...]uint8{0, 30, 54, 84, 108, 138, 162}
)

func (i ContingentCondition) String() string {
	switch {
	case 49 <= i && i <= 57:
		i -= 49
		return _ContingentCondition_name_0[_ContingentCondition_index_0[i]:_ContingentCondition_index_0[i+1]]
	case i == 72:
		return _ContingentCondition_name_1
	case 74 <= i && i <= 79:
		i -= 74
		return _ContingentCondition_name_2[_ContingentCondition_index_2[i]:_ContingentCondition_index_2[i+1]]
	case i == 90:
		return _ContingentCondition_name_3
	default:
		return "ContingentCondition(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}

// Code generated by "stringer -type OrderStatus -linecomment"; DO NOT EDIT.

package rhmonitor4go

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[RH_TRADE_OST_AllTraded-48]
	_ = x[RH_TRADE_OST_PartTradedQueueing-49]
	_ = x[RH_TRADE_OST_PartTradedNotQueueing-50]
	_ = x[RH_TRADE_OST_NoTradeQueueing-51]
	_ = x[RH_TRADE_OST_NoTradeNotQueueing-52]
	_ = x[RH_TRADE_OST_Canceled-53]
	_ = x[RH_TRADE_OST_Unknown-97]
	_ = x[RH_TRADE_OST_NotTouched-98]
	_ = x[RH_TRADE_OST_Touched-99]
	_ = x[RH_TRADE_OST_Submitted-100]
	_ = x[RH_TRADE_OST_Amending-109]
}

const (
	_OrderStatus_name_0 = "全部成交部分成交还在队列中部分成交不在队列中未成交还在队列中未成交不在队列中撤单"
	_OrderStatus_name_1 = "未知尚未触发已触发已提交"
	_OrderStatus_name_2 = "正在修改"
)

var (
	_OrderStatus_index_0 = [...]uint8{0, 12, 39, 66, 90, 114, 120}
	_OrderStatus_index_1 = [...]uint8{0, 6, 18, 27, 36}
)

func (i OrderStatus) String() string {
	switch {
	case 48 <= i && i <= 53:
		i -= 48
		return _OrderStatus_name_0[_OrderStatus_index_0[i]:_OrderStatus_index_0[i+1]]
	case 97 <= i && i <= 100:
		i -= 97
		return _OrderStatus_name_1[_OrderStatus_index_1[i]:_OrderStatus_index_1[i+1]]
	case i == 109:
		return _OrderStatus_name_2
	default:
		return "OrderStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}

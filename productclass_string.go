// Code generated by "stringer -type ProductClass -linecomment"; DO NOT EDIT.

package rhmonitor4go

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Futures-49]
	_ = x[Options-50]
	_ = x[Combination-51]
	_ = x[Spot-52]
	_ = x[EFP-53]
	_ = x[SpotOptions-54]
	_ = x[TAS-55]
	_ = x[Stock-56]
	_ = x[StockIndex-57]
	_ = x[AimStatus-58]
	_ = x[ETFOptions-101]
	_ = x[MI-73]
}

const (
	_ProductClass_name_0 = "期货期货期权组合现货期转现现货期权TAS合约证券证券指数标的融资融券状态"
	_ProductClass_name_1 = "金属指数"
	_ProductClass_name_2 = "ETF期权"
)

var (
	_ProductClass_index_0 = [...]uint8{0, 6, 18, 24, 30, 39, 51, 60, 66, 78, 102}
)

func (i ProductClass) String() string {
	switch {
	case 49 <= i && i <= 58:
		i -= 49
		return _ProductClass_name_0[_ProductClass_index_0[i]:_ProductClass_index_0[i+1]]
	case i == 73:
		return _ProductClass_name_1
	case i == 101:
		return _ProductClass_name_2
	default:
		return "ProductClass(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}

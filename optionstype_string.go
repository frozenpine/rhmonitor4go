// Code generated by "stringer -type OptionsType -linecomment"; DO NOT EDIT.

package rhmonitor4go

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Call-49]
	_ = x[Put-50]
}

const _OptionsType_name = "看涨看跌"

var _OptionsType_index = [...]uint8{0, 6, 12}

func (i OptionsType) String() string {
	i -= 49
	if i >= OptionsType(len(_OptionsType_index)-1) {
		return "OptionsType(" + strconv.FormatInt(int64(i+49), 10) + ")"
	}
	return _OptionsType_name[_OptionsType_index[i]:_OptionsType_index[i+1]]
}
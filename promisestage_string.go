// Code generated by "stringer -type PromiseStage -linecomment"; DO NOT EDIT.

package rhmonitor4go

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[PromiseInfight-0]
	_ = x[PromiseAwait-1]
	_ = x[PromiseThen-2]
	_ = x[PromiseCatch-3]
	_ = x[PromiseFinal-4]
}

const _PromiseStage_name = "InflightAwaitThenCatchFinal"

var _PromiseStage_index = [...]uint8{0, 8, 13, 17, 22, 27}

func (i PromiseStage) String() string {
	if i >= PromiseStage(len(_PromiseStage_index)-1) {
		return "PromiseStage(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _PromiseStage_name[_PromiseStage_index[i]:_PromiseStage_index[i+1]]
}

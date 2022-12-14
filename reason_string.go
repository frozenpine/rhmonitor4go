// Code generated by "stringer -type Reason -linecomment"; DO NOT EDIT.

package rhmonitor4go

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[NetReadFailed-4097]
	_ = x[NetWriteFailed-4098]
	_ = x[HBTimeout-8193]
	_ = x[HBSendFaild-8194]
	_ = x[InvalidPacket-8195]
}

const (
	_Reason_name_0 = "网络读失败网络写失败"
	_Reason_name_1 = "接收心跳超时发送心跳失败收到错误报文"
)

var (
	_Reason_index_0 = [...]uint8{0, 15, 30}
	_Reason_index_1 = [...]uint8{0, 18, 36, 54}
)

func (i Reason) String() string {
	switch {
	case 4097 <= i && i <= 4098:
		i -= 4097
		return _Reason_name_0[_Reason_index_0[i]:_Reason_index_0[i+1]]
	case 8193 <= i && i <= 8195:
		i -= 8193
		return _Reason_name_1[_Reason_index_1[i]:_Reason_index_1[i+1]]
	default:
		return "Reason(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}

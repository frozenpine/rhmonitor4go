// Code generated by "stringer -type PrivilegeType -linecomment"; DO NOT EDIT.

package rhmonitor4go

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[RH_MONITOR_ADMINISTRATOR-48]
	_ = x[RH_MONITOR_NOMAL-49]
}

const _PrivilegeType_name = "管理员普通用户"

var _PrivilegeType_index = [...]uint8{0, 9, 21}

func (i PrivilegeType) String() string {
	i -= 48
	if i >= PrivilegeType(len(_PrivilegeType_index)-1) {
		return "PrivilegeType(" + strconv.FormatInt(int64(i+48), 10) + ")"
	}
	return _PrivilegeType_name[_PrivilegeType_index[i]:_PrivilegeType_index[i+1]]
}
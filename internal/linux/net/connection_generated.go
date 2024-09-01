// Code generated by "stringer -type=connState,connIcon -output connection_generated.go -linecomment"; DO NOT EDIT.

package net

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[connUnknown-0]
	_ = x[connActivating-1]
	_ = x[connOnline-2]
	_ = x[connDeactivating-3]
	_ = x[connOffline-4]
}

const _connState_name = "UnknownActivatingOnlineDeactivatingOffline"

var _connState_index = [...]uint8{0, 7, 17, 23, 35, 42}

func (i connState) String() string {
	if i >= connState(len(_connState_index)-1) {
		return "connState(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _connState_name[_connState_index[i]:_connState_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[iconUnknown-0]
	_ = x[iconActivating-1]
	_ = x[iconOnline-2]
	_ = x[iconDeactivating-3]
	_ = x[iconOffline-4]
}

const _connIcon_name = "mdi:help-networkmdi:plus-networkmdi:networkmdi:network-minusmdi:network-off"

var _connIcon_index = [...]uint8{0, 16, 32, 43, 60, 75}

func (i connIcon) String() string {
	if i >= connIcon(len(_connIcon_index)-1) {
		return "connIcon(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _connIcon_name[_connIcon_index[i]:_connIcon_index[i+1]]
}

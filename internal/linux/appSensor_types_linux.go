// Code generated by "stringer -type=appSensorType -output appSensor_types_linux.go -linecomment"; DO NOT EDIT.

package linux

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[activeApp-4]
	_ = x[runningApps-5]
}

const _appSensorType_name = "Active AppRunning Apps"

var _appSensorType_index = [...]uint8{0, 10, 22}

func (i appSensorType) String() string {
	i -= 4
	if i < 0 || i >= appSensorType(len(_appSensorType_index)-1) {
		return "appSensorType(" + strconv.FormatInt(int64(i+4), 10) + ")"
	}
	return _appSensorType_name[_appSensorType_index[i]:_appSensorType_index[i+1]]
}

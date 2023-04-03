// Code generated by "stringer -type=batteryProp -output battery_linux_props.go -trimprefix battery"; DO NOT EDIT.

package device

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Percentage-4]
	_ = x[Temperature-5]
	_ = x[EnergyRate-6]
	_ = x[Voltage-7]
	_ = x[Energy-8]
	_ = x[batteryState-9]
	_ = x[NativePath-10]
	_ = x[batteryType-11]
}

const _batteryProp_name = "PercentageTemperatureEnergyRateVoltageEnergyStateNativePathType"

var _batteryProp_index = [...]uint8{0, 10, 21, 31, 38, 44, 49, 59, 63}

func (i batteryProp) String() string {
	i -= 4
	if i < 0 || i >= batteryProp(len(_batteryProp_index)-1) {
		return "batteryProp(" + strconv.FormatInt(int64(i+4), 10) + ")"
	}
	return _batteryProp_name[_batteryProp_index[i]:_batteryProp_index[i+1]]
}

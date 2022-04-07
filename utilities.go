// Package kong
package kong

func valueOrDefaultValue(value interface{}, defaultValue interface{}) interface{} {
	switch value.(type) {
	case string:
		if value == "" {
			value = defaultValue
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		if value == 0 {
			value = defaultValue
		}
	case bool:
		if value == false {
			value = defaultValue
		}
	}

	return value
}

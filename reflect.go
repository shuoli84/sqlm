package sqlm

func deRef(i interface{}) interface{} {
	switch t := i.(type) {
	case *string:
		return *t
	case *int:
		return *t
	case *int8:
		return *t
	case *int16:
		return *t
	case *int32:
		return *t
	case *int64:
		return *t
	case *float32:
		return *t
	case *float64:
		return *t
	case *interface{}:
		return *t
	default:
		return t
	}
}

func flat(i interface{}) []interface{} {
	result := []interface{}{}
	switch t := i.(type) {
	case []interface{}:
		for _, e := range t {
			result = append(result, flat(e)...)
		}
	default:
		result = append(result, t)
	}

	return result
}

func assign(target interface{}, value interface{}) error {
	switch t := target.(type) {
	case *string:
		*t = deRef(value).(string)
	case *int:
		*t = deRef(value).(int)
	case *int8:
		*t = deRef(value).(int8)
	case *int16:
		*t = deRef(value).(int16)
	case *int32:
		*t = deRef(value).(int32)
	case *int64:
		*t = deRef(value).(int64)
	case *float32:
		*t = deRef(value).(float32)
	case *float64:
		*t = deRef(value).(float64)
	case *interface{}:
		*t = deRef(value)
	}

	return nil
}
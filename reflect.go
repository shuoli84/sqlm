package sqlm

import (
	"reflect"
)

var derefCount int64 = 0
var flatCount int64 = 0

func deRef(i interface{}) interface{} {
	if i == nil {
		return i
	}
	derefCount += 1
	typeOfI := reflect.TypeOf(i)
	switch typeOfI.Kind() {
	case reflect.Ptr:
		return deRef(reflect.ValueOf(i).Elem().Interface())
	default:
		return i
	}
}

func flat(i interface{}) []interface{} {
	flatCount += 1

	kindOfI := reflect.TypeOf(i).Kind()

	switch kindOfI {
	case reflect.Slice, reflect.Array:
		valueOfI := reflect.ValueOf(i)

		result := reflect.ValueOf(make([]interface{}, 0))
		// Iterate the slice and flat each of them
		for index := 0; index < valueOfI.Len(); index++ {
			v := valueOfI.Index(index)
			if v.Kind() == reflect.Interface {
				vElem := v.Elem()
				if vElem.Kind() == reflect.Slice || vElem.Kind() == reflect.Array {

					for internalIndex := 0; internalIndex < vElem.Len(); internalIndex++ {
						internalElem := v.Elem().Index(internalIndex)
						eKind := internalElem.Kind()
						if eKind == reflect.Interface && (
							internalElem.Elem().Kind() == reflect.Slice ||
							internalElem.Elem().Kind() == reflect.Array) {
							result = reflect.AppendSlice(result,
								reflect.ValueOf(flat(internalElem.Interface())))
						} else {
							result = reflect.Append(result,
								reflect.ValueOf(internalElem.Interface()))
						}
					}
				} else {
					result = reflect.Append(result, v)
				}
			} else {
				result = reflect.Append(result, v)
			}
		}
		back := result.Interface()
		return back.([]interface{})
	default:
		return []interface{}{i}
	}
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

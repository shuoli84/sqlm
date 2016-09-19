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

// flat v into list, and return list value
func flatInto(v reflect.Value, result reflect.Value) reflect.Value {
	kindOfI := v.Kind()

	switch kindOfI {
	case reflect.Slice, reflect.Array:
		vLen := v.Len()
		if vLen == 1 {
			vItem := v.Index(0)
			if vItem.Kind() == reflect.Interface {
				vElem := vItem.Elem()
				vElemKind := vElem.Kind()
				if vElemKind == reflect.Slice || vElemKind == reflect.Array {
					for internalIndex := 0; internalIndex < vElem.Len(); internalIndex++ {
						result = flatInto(vElem.Index(internalIndex), result)
					}
				} else {
					result = reflect.Append(result, vItem)
				}
			} else {
				result = reflect.Append(result, vItem)
			}
			return result
		}

		for index := 0; index < vLen; index++ {
			vItem := v.Index(index)
			if vItem.Kind() == reflect.Interface {
				vElem := vItem.Elem()
				if vElem.Kind() == reflect.Slice || vElem.Kind() == reflect.Array {
					for internalIndex := 0; internalIndex < vElem.Len(); internalIndex++ {
						internalElem := vItem.Elem().Index(internalIndex)
						result = flatInto(internalElem, result)
					}
				} else {
					result = reflect.Append(result, vItem)
				}
			} else {
				result = reflect.Append(result, vItem)
			}
		}
		return result
	default:
		result = reflect.Append(result, v)
		return result
	}
}

func flat(list []interface{}, i interface{}) []interface{} {
	flatCount += 1

	kindOfI := reflect.TypeOf(i).Kind()

	switch kindOfI {
	case reflect.Slice, reflect.Array:
		valueOfI := reflect.ValueOf(i)
		result := reflect.ValueOf(list)
		result = flatInto(valueOfI, result)
		return result.Interface().([]interface{})
	default:
		return append(list, i)
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

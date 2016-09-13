package sqlm

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

type Fielder interface {
	FieldForName(name string) interface{}
}

type Mapper struct {
	fieldNames []string
}

func NewMapper(fields []string) Mapper {
	return Mapper{fields}
}

var EmptyMapper = Mapper{}

func (r Mapper) Columns() []string {
	return r.fieldNames
}

func (r Mapper) ColumnString() string {
	return strings.Join(r.Columns(), ", ")
}

func contains(items []string, target string) bool {
	for i := 0; i < len(items); i++ {
		if items[i] == target {
			return true
		}
	}
	return false
}

// split set "from" into two sets, in and out
func split(from []string, set []string) ([]string, []string) {
	inSet := make([]string, 0)
	outSet := make([]string, 0)
	for i := 0; i < len(from); i++ {
		item := from[i]
		if contains(set, item) {
			inSet = append(inSet, item)
		} else {
			outSet = append(outSet, item)
		}
	}

	return inSet, outSet
}

func dedup(items []string) []string {
	result := []string{}
	for _, item := range(items) {
		var shouldAppend = true
		for _, existedItem := range(result) {
			if item == existedItem {
				shouldAppend = false
				break
			}
		}

		if shouldAppend {
			result = append(result, item)
		}
	}

	return result
}

// Split the mapper into two, one in the fields, one not
func (r Mapper) Split(fields []string) (Mapper, Mapper) {
	columnsIn, columnsOut := split(r.Columns(), fields)
	return NewMapper(columnsIn), NewMapper(columnsOut)
}

func (r Mapper) Append(fields ...string) Mapper {
	return NewMapper(dedup(append(r.Columns(), fields...)))
}

func (r Mapper) Fields(fielder Fielder) []interface{} {
	fields := make([]interface{}, len(r.fieldNames))
	for i, fieldName := range r.fieldNames {
		fields[i] = fielder.FieldForName(fieldName)
	}

	return fields
}

func (r Mapper) FormatSQLInsertValues(fielders []Fielder) Expression {
	valueString := bytes.Buffer{}
	args := []interface{}{}

	for index := 0; index < len(fielders); index++ {
		valueString.WriteString("(")

		for i, fieldName := range r.fieldNames {
			field := fielders[index].FieldForName(fieldName)
			if sqlShouldEscape(field) {
				valueString.WriteString("?")
				args = append(args, field)
			} else {
				valueString.WriteString(fmt.Sprintf("%v", deRef(field)))
			}

			if i < len(r.fieldNames)-1 {
				valueString.WriteString(",")
			} else {
				valueString.WriteString(")")
			}
		}

		if index < len(fielders)-1 {
			valueString.WriteString(",")
		}
	}

	return NewRaw(valueString.String(), args)
}

func (r Mapper) FormatSQLUpdateSets(fielder Fielder) Expression {
	fields := r.Fields(fielder)
	buf := bytes.Buffer{}
	args := []interface{}{}
	for _, fieldName := range r.fieldNames {
		buf.WriteString(fieldName + "=")
		field := fielder.FieldForName(fieldName)
		if sqlShouldEscape(field) {
			buf.WriteString("?")
			args = append(args, field)
		} else {
			buf.WriteString(fmt.Sprintf("%v", field))
		}
	}
	return NewRaw(string(buf.Bytes()), fields)
}

func sqlShouldEscape(i interface{}) bool {
	if _, ok := i.(string); ok {
		return true
	} else if _, ok := i.(*string); ok {
		return true
	}

	if _, ok := i.(time.Time); ok {
		return true
	} else if _, ok := i.(*time.Time); ok {
		return true
	}

	if _, ok := i.([]byte); ok {
		return true
	} else if _, ok := i.(*[]byte); ok {
		return true
	}

	return false
}

func (r Mapper) PackDict(fielder Fielder) map[string]interface{} {
	result := map[string]interface{}{}

	for _, fieldName := range r.fieldNames {
		result[fieldName] = fielder.FieldForName(fieldName)
	}

	return result
}

// Load from dict, you can treat it as a way of deserialize, but just from map
func (r Mapper) LoadFromDict(dict map[string]interface{}, fielder Fielder) {
	for _, fieldName := range r.fieldNames {
		if v, ok := dict[fieldName]; ok {
			assign(fielder.FieldForName(fieldName), v)
		} else {
			continue
		}
	}
}

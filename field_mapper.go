package sqlm

import (
	"bytes"
	"strings"
	"fmt"
)

type Fielder interface {
	FieldForName(name string) interface{}
}

type Mapper struct {
	fieldNames []string
}

func NewMapper(fields []string) *Mapper {
	return &Mapper{fields}
}

func (r *Mapper) Columns() []string {
	return r.fieldNames
}

func (r *Mapper) ColumnString() string {
	return strings.Join(r.Columns(), ", ")
}

func (r *Mapper) Fields(fielder Fielder) []interface{} {
	fields := make([]interface{}, len(r.fieldNames))
	for i, fieldName := range r.fieldNames {
		fields[i] = fielder.FieldForName(fieldName)
	}

	return fields
}

func (r *Mapper) FormatSQLInsertValues(fielders []Fielder) (string, []interface{}) {
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

			if i < len(r.fieldNames) - 1 {
				valueString.WriteString(",")
			} else {
				valueString.WriteString(")")
			}
		}

		if index < len(fielders) - 1 {
			valueString.WriteString(",")
		}
	}

	return valueString.String(), args
}

func (r *Mapper) FormatSQLUpdateSets(fielder Fielder) (string, []interface{}) {
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
	return string(buf.Bytes()), fields
}

func sqlShouldEscape(i interface{}) bool {
	if _, ok := i.(string); ok {
		return true
	} else if _, ok := i.(*string); ok {
		return true
	}
	return false
}

func (r *Mapper) PackDict(fielder Fielder) map[string]interface{} {
	result := map[string]interface{}{}

	for _, fieldName := range r.fieldNames {
		result[fieldName] = fielder.FieldForName(fieldName)
	}

	return result
}

// Load from dict, you can treat it as a way of deserialize, but just from map
func (r *Mapper) LoadFromDict(dict map[string]interface{}, fielder Fielder) {
	for _, fieldName := range r.fieldNames {
		if v, ok := dict[fieldName]; ok {
			assign(fielder.FieldForName(fieldName), v)
		} else {
			continue
		}
	}
}

type FielderMap map[string]interface{}

func (r *FielderMap) FieldForName(name string) interface{} {
	return map[string]interface{}(*r)[name]
}

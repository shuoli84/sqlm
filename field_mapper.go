package sqlm

import (
	"bytes"
	"strings"
)

type Fielder interface {
	FieldForName(name string) interface{}
}

type FieldsMapper struct {
	fieldNames []string
}

func NewFieldsMapper(fields []string) *FieldsMapper {
	return &FieldsMapper{fields}
}

func (r *FieldsMapper) Columns() []string {
	return r.fieldNames
}

func (r *FieldsMapper) ColumnString() string {
	return strings.Join(r.Columns(), ", ")
}

func (r *FieldsMapper) Fields(fielder Fielder) []interface{} {
	fields := make([]interface{}, len(r.fieldNames))
	for i, fieldName := range r.fieldNames {
		fields[i] = fielder.FieldForName(fieldName)
	}

	return fields
}

func (r *FieldsMapper) MapFields(fielders []Fielder) []interface{} {
	result := []interface{} {}
	for _, fielder := range fielders {
		result = append(result, r.Fields(fielder)...)
	}

	return result
}

func (r *FieldsMapper) ValuesPlaceholder(count int) string {
	valueString := bytes.Buffer{}

	for index := 0; index < count; index++ {
		valueString.WriteString("(")

		for i, _ := range r.fieldNames {
			valueString.WriteString("?")
			if i < len(r.fieldNames) - 1 {
				valueString.WriteString(",")
			} else {
				valueString.WriteString(")")
			}
		}

		if index < count - 1 {
			valueString.WriteString(", ")
		}
	}

	return valueString.String()
}

func (r *FieldsMapper) PackDict(fielder Fielder) map[string]interface{} {
	result := map[string]interface{}{}

	for _, fieldName := range r.fieldNames {
		result[fieldName] = fielder.FieldForName(fieldName)
	}

	return result
}

type FielderMap map[string]interface{}

func (r *FielderMap) FieldForName(name string) interface{} {
	return map[string]interface{}(*r)[name]
}

package sqlm

import (
	"fmt"
	"testing"
)

type TestFielder2 struct {
	Field1 string
	Field2 string
	Field3 int
}

func (r *TestFielder2) FieldForName(name string) interface{} {
	switch name {
	case "field_1":
		return &r.Field1
	case "field_2":
		return &r.Field2
	case "field_3":
		return &r.Field3
	}

	return nil
}

func TestFieldMapper(t *testing.T) {
	mapper := NewMapper([]string{"field_1", "field_2", "field_3"})
	values := []Fielder{
		&TestFielder2{"1", "2", 1},
		&TestFielder2{"3", "4", 2},
		&TestFielder2{"5", "6", 3},
		&TestFielder2{"7", "8", 4},
		&TestFielder2{"9", "10", 5},
	}

	sql, args := mapper.FormatSQLInsertValues(values).ToSql()
	fmt.Println(sql)
	fmt.Printf("%v\n", args)

	dict := mapper.PackDict(values[0])
	fmt.Printf("%v\n", dict)

	v := &TestFielder2{}
	mapper.LoadFromDict(dict, v)
}

package sqlm

import (
	"strings"
	"fmt"
	"bytes"
)

type Expression interface {
	Sql() (string, []interface{})
}

type Raw struct {
	value string
	arguments []interface{}
}

func (s Raw) Sql() (string, []interface{}) {
	return s.value, s.arguments
}

func SetMapper(mapper *FieldsMapper, fielder Fielder) Expression {
	columns := mapper.Columns()
	fields := mapper.Fields(fielder)

	buf := bytes.Buffer{}
	for i := 0; i < len(columns); i++ {
		if i < len(columns) - 1 {
			buf.WriteString(fmt.Sprintf("%s = ?,", columns[i]))
		} else {
			buf.WriteString(fmt.Sprintf("%s = ?", columns[i]))
		}
	}

	return Raw {
		value: fmt.Sprintf("SET %s", string(buf.String())),
		arguments: fields,
	}
}

type joiner struct {
	expressions []interface{}
	sep string
	prefix string
	suffix string
}

func (s joiner) Sql() (string, []interface{}) {
	sql := []string{}
	arguments := []interface{}{}
	for _, exp := range s.expressions {
		if exp != nil {
			expSql, expArgs := Exp(exp).Sql()
			sql = append(sql, expSql)
			arguments = append(arguments, expArgs...)
		}
	}

	return s.prefix + strings.Join(sql, s.sep) + s.suffix, arguments
}

func G(components ...interface{}) Expression {
	return Exp("(", components, ")")
}

func And(expressions ...interface{}) Expression {
	return joiner{
		expressions: expressions,
		sep: " AND ",
		prefix: "(",
		suffix: ")",
	}
}

func Or(filters ...interface{}) Expression {
	return joiner{
		expressions: filters,
		sep: " OR ",
		prefix: "(",
		suffix: ")",
	}
}

func Not(exp interface{}) Expression {
	return Exp("NOT", exp)
}

func flat(expressions ...interface{}) []interface{} {
	result := []interface{}{}
	for _, e := range expressions {
		switch t := e.(type) {
		case []interface{}:
			result = append(result, flat(t...)...)
		default:
			result = append(result, t)
		}
	}

	return result
}

func Join(sep string, expressions ...interface{}) Expression {
	return joiner{
		expressions: flat(expressions),
		sep: sep,
	}
}

func Build(expressions ...interface{}) (string, []interface{}) {
	return Exp(expressions).Sql()
}

type Param struct{
	inner interface{}
}

func P(value interface{}) Param {
	return Param{inner: value}
}

type Value struct {
	inner interface{}
}

func V(values ...interface{}) []Value {
	result := make([]Value, 0, len(values))
	for i := 0; i < len(values); i++ {
		result = append(result, Value{inner: values[i]})
	}
	return result
}

func Exp(components ...interface{}) Expression {
	// If the component is param then we wrap it
	// Otherwise, we just append it to sql expression
	buf := bytes.Buffer{}
	arguments := []interface{}{}
	for _, c := range components {
		if p, ok := c.(Param); ok {
			buf.WriteString(" ? ")
			arguments = append(arguments, p.inner)
		} else if values, ok := c.([]Value); ok {
			for index, v := range values {
				if index < len(values) - 1 {
					buf.WriteString(fmt.Sprintf("%v, ", deRef(v.inner)))
				} else {
					buf.WriteString(fmt.Sprintf(" %v", deRef(v.inner)))
				}
			}
		} else if p, ok := c.(Expression); ok {
			sql, args := p.Sql()
			buf.WriteString(sql)
			buf.WriteRune(' ')
			arguments = append(arguments, args...)
		} else if expressions, ok := c.([]Expression); ok {
			for index, exp := range expressions {
				sql, args := exp.Sql()
				if index < len(values) - 1 {
					buf.WriteString(fmt.Sprintf("%v ", sql))
				} else {
					buf.WriteString(fmt.Sprintf(" %v", sql))
				}
				arguments = append(arguments, args...)
			}
		} else if slice, ok := c.([]interface{}); ok {
			sql, args := Exp(slice...).Sql()
			fmt.Printf("slicing %v: %s\n", slice, sql)

			buf.WriteString(sql)
			arguments = append(arguments, args...)
		} else {
			if s, ok := c.(string); ok {
				buf.WriteString(" " + s + " ")
			} else {
				buf.WriteString(fmt.Sprintf(" %v ", deRef(c)))
			}
		}
	}

	return Raw {
		value: string(buf.Bytes()),
		arguments: arguments,
	}
}

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

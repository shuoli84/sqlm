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
		sep: " AND",
		prefix: "(",
		suffix: ")",
	}
}

func Or(filters ...interface{}) Expression {
	return joiner{
		expressions: filters,
		sep: " OR",
		prefix: "(",
		suffix: ")",
	}
}

func Not(exp interface{}) Expression {
	return Exp("NOT", exp)
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

func V(v interface{}) Value {
	return Value{inner: v}
}

func Exp(components ...interface{}) Expression {
	// If the component is param then we wrap it
	// Otherwise, we just append it to sql expression
	buf := bytes.Buffer{}
	arguments := []interface{}{}
	for _, c := range flat(components) {
		if p, ok := c.(Param); ok {
			buf.WriteString(" ?")
			arguments = append(arguments, p.inner)
		} else if v, ok := c.(Value); ok {
			buf.WriteString(fmt.Sprintf(" %v", deRef(v.inner)))
		} else if p, ok := c.(Expression); ok {
			sql, args := p.Sql()
			buf.WriteString(" " + sql)
			arguments = append(arguments, args...)
		} else {
			if s, ok := c.(string); ok {
				buf.WriteString(" " + s)
			} else {
				buf.WriteString(fmt.Sprintf(" %v", deRef(c)))
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

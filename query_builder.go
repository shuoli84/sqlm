package sqlm

import (
	"strings"
	"fmt"
	"bytes"
)

type Expression interface {
	Sql(context interface{}) (string, []interface{})
}

type Raw struct {
	sql       string
	arguments []interface{}
}

func (s Raw) Sql(context interface{}) (string, []interface{}) {
	return s.sql, s.arguments
}

func NewRaw(sql string, arguments ...interface{}) Raw {
	return Raw{sql: sql, arguments: arguments}
}

type joiner struct {
	expressions []interface{}
	sep string
	prefix string
	suffix string
}

func (s joiner) Sql(context interface{}) (string, []interface{}) {
	sql := []string{}
	arguments := []interface{}{}
	for _, exp := range s.expressions {
		if exp != nil {
			expSql, expArgs := Exp(exp).Sql(context)
			sql = append(sql, expSql)
			arguments = append(arguments, expArgs...)
		}
	}

	return s.prefix + strings.Join(sql, s.sep) + s.suffix, arguments
}

type node struct {
	expressions []Expression
}

func (r node) Sql(context interface{}) (string, []interface{}) {
	buf := bytes.Buffer{}
	arguments := []interface{}{}

	for _, e := range r.expressions {
		var expSql string
		var expArgs []interface{}

		// If we have a context by our self, then we pass it down
		// Otherwise, pass the passed in context down
		expSql, expArgs = e.Sql(context)

		buf.WriteString(" " + expSql)
		arguments = append(arguments, expArgs...)
	}

	return buf.String(), arguments
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
		// When expression passed in as [[1,2,3]], we prefer it converts to [1,2,3]
		expressions: flat(expressions),
		sep: sep,
	}
}

func Build(expressions ...interface{}) (string, []interface{}) {
	return Exp(expressions).Sql(nil)
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
	expressions := []Expression{}
	components = flat(components)

	toBeMerged := []string{}
	shouldDoMerge := false

	for i := 0; i < len(components); i++ {
		c := components[i]

		var exp Expression
		switch v := c.(type) {
		case Param:
			exp = NewRaw("?", v.inner)
			shouldDoMerge = true
		case Value:
			toBeMerged = append(toBeMerged, fmt.Sprintf("%v", deRef(v.inner)))
		case Expression:
			exp = v
			shouldDoMerge = true
		case string:
			toBeMerged = append(toBeMerged, v)
			shouldDoMerge = false
		default:
			toBeMerged = append(toBeMerged, fmt.Sprintf("%v", deRef(v)))
			shouldDoMerge = false
		}

		if shouldDoMerge {
			expressions = append(expressions, NewRaw(strings.Join(toBeMerged, " ")))
			toBeMerged = []string{}
			shouldDoMerge = false
		}

		if exp != nil {
			expressions = append(expressions, exp)
		}
	}
	if len(toBeMerged) > 0 {
		expressions = append(expressions, NewRaw(strings.Join(toBeMerged, " ")))
	}

	return node{expressions: expressions}
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

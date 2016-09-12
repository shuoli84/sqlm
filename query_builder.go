package sqlm

import (
	"fmt"
	"strings"
)

// Any thing can be converted to a sql and its arguments is an expression
type Expression interface {
	ToSql() (string, []interface{})
}

// Raw expression just a wrapper of a sql and relative arguments
type Raw struct {
	Sql       string
	Arguments []interface{}
}

func (s Raw) ToSql() (string, []interface{}) {
	return s.Sql, s.Arguments
}

func NewRaw(sql string, arguments ...interface{}) Raw {
	return Raw{Sql: sql, Arguments: arguments}
}

// formatter is a generic helper, which provide the way to join several expressions
// together.
type formatter struct {
	expressions []Expression
	sep         string
	prefix      string
	suffix      string
}

func (s formatter) ToSql() (string, []interface{}) {
	sql := make([]string, len(s.expressions))
	arguments := make([]interface{}, 0, len(s.expressions))
	for i, expression := range s.expressions {
		expSql, expArgs := expression.ToSql()
		sql[i] = expSql
		arguments = append(arguments, expArgs...)
	}

	return s.prefix + strings.Join(sql, s.sep) + s.suffix, arguments
}

func G(components ...interface{}) Expression {
	return F("(1 2)", components...)
}

func And(components ...interface{}) Expression {
	return F("(1 AND 2)", components...)
}

func Or(components ...interface{}) Expression {
	return F("(1 OR 2)", components...)
}

func Not(exp interface{}) Expression {
	return F("NOT 12", exp)
}

// sep format, like dateformatter, we use magic numbers to split
// (1,2) => prefix:( sep:, suffix:)
// 1,2  => prefix: sep:, suffix:
// If the sep has three letters, then the first is prefix, last is suffix and middle is the sep

func F(sepFormat string, expressions ...interface{}) Expression {
	var prefix, sep, suffix string

	components := strings.Split(sepFormat, "1")
	prefix = components[0]

	secondHalfComponents := strings.Split(components[1], "2")
	sep = secondHalfComponents[0]
	suffix = secondHalfComponents[1]

	return formatter{
		// When expression passed in as [[1,2,3]], we prefer it converts to [1,2,3]
		expressions: componentsToExpressions(expressions),
		prefix:      prefix,
		sep:         sep,
		suffix:      suffix,
	}
}

func Build(expressions ...interface{}) (string, []interface{}) {
	return Exp(expressions).ToSql()
}

type Param struct {
	inner interface{}
}

func P(value interface{}) Expression {
	return Raw{"?", []interface{}{value}}
}

func Exp(components ...interface{}) Expression {
	return F("1 2", components)
}

func componentsToExpressions(components []interface{}) []Expression {
	expressions := []Expression{}
	components = flat(components)

	for i := 0; i < len(components); i++ {
		c := components[i]

		var exp Expression
		switch v := c.(type) {
		case Expression:
			exp = v
		case string:
			exp = NewRaw(v)
		default:
			exp = NewRaw(fmt.Sprintf("%v", deRef(v)))
		}

		expressions = append(expressions, exp)
	}

	return expressions
}

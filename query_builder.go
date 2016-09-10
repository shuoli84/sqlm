package sqlm

import (
	"strings"
	"fmt"
	"bytes"
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
	components []interface{}
	sep        string
	prefix     string
	suffix     string
}

func (s formatter) ToSql() (string, []interface{}) {
	sql := make([]string, len(s.components))
	arguments := make([]interface{}, 0, len(s.components))
	for i, component := range s.components {
		expSql, expArgs := Exp(component).ToSql()
		sql[i] = expSql
		arguments = append(arguments, expArgs...)
	}

	return s.prefix + strings.Join(sql, s.sep) + s.suffix, arguments
}

// node holds several expressions and join them one by one
type node struct {
	expressions []Expression
}

func (r node) ToSql() (string, []interface{}) {
	buf := bytes.Buffer{}
	arguments := []interface{}{}

	for _, e := range r.expressions {
		var expSql string
		var expArgs []interface{}

		expSql, expArgs = e.ToSql()

		buf.WriteString(" " + expSql)
		arguments = append(arguments, expArgs...)
	}

	return buf.String(), arguments
}

func G(components ...interface{}) Expression {
	return formatter{
		components: components,
		prefix: "(",
		suffix: ")",
	}
}

func And(expressions ...interface{}) Expression {
	return formatter{
		components: expressions,
		sep: " AND",
		prefix: "(",
		suffix: ")",
	}
}

func Or(filters ...interface{}) Expression {
	return formatter{
		components: filters,
		sep: " OR",
		prefix: "(",
		suffix: ")",
	}
}

func Not(exp interface{}) Expression {
	return formatter{
		components: []interface{}{exp},
		prefix: "NOT",
	}
}


// sep format
// (,)  (1,2,3,4)
// , 1,2,3,4
// If the sep has three letters, then the first is prefix, last is suffix and middle is the sep

func Format(sepFormat string, expressions ...interface{}) Expression {
	var prefix, sep, suffix string

	if len(sepFormat) == 3 {
		components := strings.Split(sepFormat, "")
		prefix, sep, suffix = components[0], components[1], components[2]
	} else {
		sep = sepFormat
	}

	return formatter{
		// When expression passed in as [[1,2,3]], we prefer it converts to [1,2,3]
		components: flat(expressions),
		prefix: prefix,
		sep: sep,
		suffix: suffix,
	}
}

func Build(expressions ...interface{}) (string, []interface{}) {
	return Exp(expressions).ToSql()
}

type Param struct{
	inner interface{}
}

func P(value interface{}) Param {
	return Param{inner: value}
}

type value struct {
	inner interface{}
}

func V(v interface{}) value {
	return value{inner: v}
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
		case value:
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


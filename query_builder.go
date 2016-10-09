package sqlm

import (
	"fmt"
	"strings"
	"time"
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
	if len(arguments) == 0 {
		return Raw{Sql: sql, Arguments: arguments}
	} else if len(arguments) == 1 {
		return Raw{Sql: sql, Arguments: flat([]interface{}{}, arguments[0])}
	}

	return Raw{Sql: sql, Arguments: flat([]interface{}{}, arguments)}
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

	sql[0] = s.prefix + sql[0]
	sql[len(sql) - 1] = sql[len(sql) - 1] + s.suffix

	return strings.Join(sql, s.sep), arguments
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
func F(sepFormat string, components ...interface{}) Expression {
	var prefix, sep, suffix string

	formatComponents := strings.Split(sepFormat, "1")
	prefix = formatComponents[0]

	secondHalfComponents := strings.Split(formatComponents[1], "2")
	sep = secondHalfComponents[0]
	suffix = secondHalfComponents[1]

	return formatter{
		// When expression passed in as [[1,2,3]], we prefer it converts to [1,2,3]
		expressions: componentsToExpressions(components),
		prefix:      prefix,
		sep:         sep,
		suffix:      suffix,
	}
}

func Build(expressions ...interface{}) (string, []interface{}) {
	sql, args := Exp(expressions...).ToSql()
	return sql, args
}

// Convert all components to Value expression
// E.g, 1 => 1
//      "what" => "?" args: "what"
//      Time => "?" args: "Time"
func P(components ...interface{}) []Expression {
	components = flat([]interface{}{}, components)
	expressions := make([]Expression, len(components))

	for i := 0; i < len(components); i++ {
		c := components[i]

		var exp Expression
		switch v := c.(type) {
		case Expression:
			exp = v
		case string, *string, []byte, time.Time, *[]byte, *time.Time:
			exp = NewRaw("?", v)
		default:
			exp = NewRaw(fmt.Sprintf("%v", deRef(v)))
		}

		expressions[i] = exp
	}
	return expressions
}

// Exp("SELECT", "a, b", "FROM", tableName) => "SELECT a, b FROM table". Use space to join all expressions
func Exp(components ...interface{}) Expression {
	return F("1 2", components...)
}

// Apply to non-value sql expression. convert arbitrary types to string and arguments
func componentsToExpressions(components []interface{}) []Expression {
	expressions := []Expression{}
	for _, component := range components {
		if v, ok := component.([]Expression); ok {
			expressions = append(expressions, v...)
		} else {
			flatted := flat(make([]interface{}, 0), component)

			for i := 0; i < len(flatted); i++ {
				c := flatted[i]

				var exp Expression
				switch v := c.(type) {
				case Expression:
					exp = v
				case string:
					exp = NewRaw(v)
				case *string:
					exp = NewRaw(*v)
				case []byte, time.Time, *[]byte, *time.Time:
					exp = NewRaw("?", c)
				default:
					exp = NewRaw(fmt.Sprintf("%v", deRef(v)))
				}

				expressions = append(expressions, exp)
			}
		}
	}

	return expressions
}

package sqlm

import (
	"strings"
	"fmt"
	"github.com/mailru/easyjson/buffer"
	"bytes"
)

type Expression interface {
	Sql() (string, []interface{})
}

type Raw struct {
	value string
	arguments []interface{}
}

func (s *Raw) Sql() (string, []interface{}) {
	return s.value, s.arguments
}

func Values(mapper *FieldsMapper, fielders []DBFielder) Expression {
	return &Raw {
		value: fmt.Sprintf(
			"(%s) VALUES %s",
			mapper.ColumnString(),
			mapper.ValuesPlaceholder(len(fielders)),
		),
		arguments: mapper.MapFields(fielders),
	}
}

func SetMapper(mapper *FieldsMapper, fielder DBFielder) Expression {
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

	return &Raw {
		value: fmt.Sprintf("SET %s", string(buf.String())),
		arguments: fields,
	}
}

func Set(expressions ...Expression) Expression {
	return Joiner {
		expressions: expressions,
		sep: ",",
		prefix: "SET ",
	}
}

type Joiner struct {
	expressions []Expression
	sep string
	prefix string
	suffix string
}

func (s Joiner) Sql() (string, []interface{}) {
	sql := []string{}
	arguments := []interface{}{}
	for _, exp := range s.expressions {
		if exp != nil {
			expSql, expArgs := exp.Sql()
			sql = append(sql, expSql)
			arguments = append(arguments, expArgs...)
		}
	}

	return s.prefix + strings.Join(sql, s.sep) + s.suffix, arguments
}

func And(expressions ...Expression) Expression {
	return Joiner {
		expressions: expressions,
		sep: " AND ",
		prefix: "(",
		suffix: ")",
	}
}

func Or(filters ...Expression) Expression {
	return Joiner{
		expressions: filters,
		sep: " OR ",
		prefix: "(",
		suffix: ")",
	}
}

func Not(exp Expression) Expression {
	return Joiner {
		expressions: []Expression {exp},
		prefix: "NOT (",
		suffix: ")",
	}
}

func Join(sep string, expressions ...Expression) Expression {
	return Joiner {
		expressions: expressions,
		sep: sep,
	}
}

func Build(expressions ...Expression) (string, []interface{}) {
	return Joiner {
		expressions: expressions,
		sep: " ",
	}.Sql()
}

type Param struct{
	inner interface{}
}

func P(values ...interface{}) []Param {
	result := make([]Param, 0, len(values))
	for i := 0; i < len(values); i++ {
		result = append(result, Param{inner: values[i]})
	}
	return result
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
	buf := buffer.Buffer{}
	arguments := []interface{}{}
	for _, c := range components {
		if params, ok := c.([]Param); ok {
			for index, p := range params {
				if index < len(params) - 1 {
					buf.AppendString("?, ")
				} else {
					buf.AppendString(" ?")
				}

				arguments = append(arguments, p.inner)
			}
		} else if p, ok := c.(Param); ok {
			buf.AppendString(" ? ")
			arguments = append(arguments, p.inner)
		} else if values, ok := c.([]Value); ok {
			for index, v := range values {
				if index < len(values) - 1 {
					buf.AppendString(fmt.Sprintf("%v, ", v.inner))
				} else {
					buf.AppendString(fmt.Sprintf(" %v", v.inner))
				}
			}
		} else if v, ok := c.(Value); ok {
			buf.AppendString(fmt.Sprintf(" %v ", v.inner))
		} else if p, ok := c.(Expression); ok {
			sql, args := p.Sql()
			buf.AppendString(sql + " ")
			arguments = append(arguments, args...)
		} else if expressions, ok := c.([]Expression); ok {
			for index, exp := range expressions {
				sql, args := exp.Sql()
				if index < len(values) - 1 {
					buf.AppendString(fmt.Sprintf("%v ", sql))
				} else {
					buf.AppendString(fmt.Sprintf(" %v", sql))
				}
				arguments = append(arguments, args...)
			}
		} else {
			if s, ok := c.(string); ok {
				buf.AppendString(s + " ")
			} else {
				buf.AppendString(fmt.Sprintf("%v ", c))
			}
		}
	}

	return &Raw {
		value: string(buf.BuildBytes()),
		arguments: arguments,
	}
}


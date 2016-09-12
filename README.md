# sqlm
A minimalist sql builder

# Motivation

* Used and tired of ORM. It adds so many abstraction and you have to learn its paradim to do something non-trivial.
* Getting tired of fmt.Sprintf of SQL, you have to put a %s and then append some value, or put a ? and counting the args number.
* Composible, easily composed by table name, columns and arguments which defined at different places. 

So here comes sqlm, which just format sql, escape the arguments with the same structure as raw sql.

```Go
db.Query(fmt.Sprintf(
	`SELECT %s FROM %s 
 	WHERE %s = ?`, columns, table, fieldName), 
 	fieldValue);
```
vs
```Go
sql, args := sqlm.Build(
	"SELECT", columns, "FROM", table,
	"WHERE", fieldName, "=", sqlm.P(fieldValue),
)
db.Query(sql, args...)
```

Examples:
Select

```Go
sql, args := sqlm.Build(
		"SELECT abc, def FROM what",
		"WHERE abc =", 1,
		"AND bcd =", sqlm.P(2),
)
```
Yields:
```SQL
SELECT abc, def FROM what WHERE abc = 1 AND bcd = ?
// args: [2]
```

Select with nested condition
```Go
	b := 23
	sql, args := sqlm.Build(
		"SELECT a, b, c, d FROM table",
		"WHERE",
		sqlm.And(
			sqlm.Or(
				sqlm.Exp("a = 3"),
				sqlm.Exp("b =", sqlm.P(b)),
			),
			sqlm.Exp("c", "=", sqlm.P(24)),
		),
	)
```
Yields:
```SQL
SELECT a, b, c, d FROM table WHERE ((a = 3 OR b = ?) AND c = ?)

// args: [23, 24]
```

# Docs
## Expression
An interface which is the building block of this lib. Other operators either generate Expressions or Transform Expressions or Composite Expressions. 
You get the idea.
```Go
type Expression interface{
    ToSql() (string, []interface{})
}
```

## sqlm.Exp
It wraps all its arguments into an Expression
```Go
sqlm.Exp("a", "=", "1")
sqlm.Exp(sqlm.Exp("a =", 1), "AND", sqlm.Exp("b = 2"))
```
## sqlm.P
This indicates that it should be treated as dynamic parameter, so the ? generated and the value put into arguments
```Go
str := ReadFromRequest()
sql, arguments := sqlm.Exp("a =", sqlm.P(str)).ToSql()
// sql == "a = ?"
// arguments == [str]
```

## sqlm.F
F stands for Format. The format defines how it composites all the expressions.

```Go
sqlm.F("(1,2)", "a = 1", "b = 2") // (a = 1,b = 2)
sqlm.F("1, 2", "a = 1", "b = 2") // a = 1, b = 2
sqlm.F("{prefix}1{sep}2{suffix}", "a = 1", "b = 2", "c = 3") // {prefix}a = 1{sep}b = 2{sep}c = 3{suffix}
```

## sqlm.And
It is just an alias of
```Go
F("(1 AND 2)", expressions...)
```

## sqlm.Or
```Go
F("(1 OR 2)", expressions...)
```

# sqlm.G
Group, alias to
```Go
F("(1 2)"
```

# WAIT, where is the struct?
Bind with struct is a nongoal for this lib. And also pls ignore the field_mapper.go. Will remove it in the future.

# Some note
Bugs expected, just like any other OOS.

# MIT LICENSE

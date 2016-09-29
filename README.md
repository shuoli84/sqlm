# sqlm
A minimalist sql builder

# Motivation

* Used and tired of ORM. It adds so many abstraction and you have to learn its paradim to do something non-trivial.
* Getting tired of fmt.Sprintf of SQL, you have to put a %s and then append some value, or put a ? and counting the args number.
* Composible, easily composed by table name, columns and arguments which defined at different places. 

So here comes sqlm, which just format sql, escape the arguments with the same structure as raw sql.

```Go
// We should only retrieve columns asked by clients, so column list is passed in
// Table name defined some where else.
// The field name is defined near the model, so rename a field won't miss a single statement.
// Parameter of int, they don't need to be argumented as ?, caz no sql injection possible for int.
// Combine all of above, here comes a mess sql statement.
db.Query(fmt.Sprintf(
	`SELECT %s FROM %s 
 	WHERE %s = ? AND %s = %d AND %s = ?`, columns, table, fieldName1, fieldName2, fieldValue2, fieldName3), 
 	fieldValue1, fieldValue3);
```
vs
```Go
// Meets all the requirements and still looks sane.
sql, args := sqlm.Build(
	"SELECT", columns, "FROM", table,
	"WHERE", sqlm.And(
		sqlm.Exp(fieldName1, "=", sqlm.P(fieldValue)),
		sqlm.Exp(fieldName2, "=", fieldValue2),
		sqlm.Exp(fieldName3, "=", sqlm.P(fieldValue3)),
	),
)
db.Query(sql, args...)
```
- sqlm.Build takes ...interface{} as input, and it concatenate them to build the sql statement.
- sqlm.P marks that the passed in variable should generate a "?" in sql statement and its value append to the argument list.
- A simple yet flexible F(ormat) method, which makes writing values list, where condition a piece of cake.

```Go
// generate some records to be inserted
recordsToBeInsert := []Record{.......}

// map each of record to sqlm.Expressions
expressions := []sqlm.Expression{} 
for _, record := range recordsToBeInsert {
    expressions = append(expressions, sqlm.F("(1, 2)", &record.field1, sqlm.P(&record.field2), &record.field3)) // (field1Value, field2Value, field3Value)
}

sqlm.Build(
	"INSERT INTO", table, sqlm.F("(1, 2)", columnList),
	"VALUES", sqlm.F("1,\n2", expressions),
)
```
- Above example is complicated, it tries to bulk insert several dynamic generated records. So we first build value expressions for each record. sqlm.F("(1, 2)", fields...) will generate sql like (field1, P(field2), field3), and field2 will be replaced by a ?. Handy!
- Then sqlm.F("1, \n2", expressions) will combine above generated expressions into 
```sql
(field1, field2, field3),  // for record1
(field1, field2, field3),  // for record2
```
- It has two parts, but still clear.

# Some other Examples:
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

# Issues
* Now when the parameter list is large, around 10000, it took around 10ms to build the sql and argument list. 

# WAIT, where is the struct?
Bind with struct is a nongoal for this lib. And also pls ignore the field_mapper.go. Will remove it in the future.

# Some note
Bugs expected, just like any other OOS.

# MIT LICENSE

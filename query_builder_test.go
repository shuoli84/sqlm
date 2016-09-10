package sqlm

import (
	"testing"
	"fmt"
	"errors"
)

type TestFielder struct {
	Field1 string
	Field2 string
}

func (t TestFielder) FieldForName(name string) interface{} {
	switch name {
	case "field1":
		return &t.Field1
	case "field2":
		return &t.Field2
	default:
		panic(errors.New("hh"))
	}
}

func TestQueryBuilder(t *testing.T) {
	{
		sql, arguments := Exp(
			"SELECT abc, def FROM what",
			"WHERE", Not(
				And(
					Exp("user_id >", P(12345)),
					And(
						Exp("media_id <", 12345),
						Exp("time_uuid =", 12345),
					),
					And(
						Exp("media_id <", 12345),
						Exp("time_uuid =", 12345),
					),
				),
			),
		).ToSql()

		fmt.Println(sql)
		fmt.Printf("len: %d, %v\n", len(arguments), arguments)
	}

	{
		i := 30
		sql, arguments := Build(
			"UPDATE table2 SET",
			F("1, \n2",
				Exp("a =", P("300")),
				Exp("b =", P("400")),
				Exp("c =", P("500")),
				Exp("d =", i),
			),
			"WHERE a =", P(300),
		)

		fmt.Println(sql)
		fmt.Printf("len: %d, %v\n", len(arguments), arguments)
	}

	{
		table := "tablename"
		sql, arguments := Exp(
			"DELETE FROM", table,
			"WHERE abc =", 1,
		).ToSql()

		fmt.Println(sql)
		fmt.Printf("len: %d, %v\n", len(arguments), arguments)
	}

	{

		// SELECT * FROM table
		// WHERE abc = 1 AND bcd = 2 AND (abc = 1  AND def >=  ?)

		sql, arguments := Build(
			"SELECT * FROM table",
			"WHERE abc =", 1, "AND", "bcd =", 2, "AND",
			F("(1 AND 2)",
				Exp("abc", "=", "1"),
				Exp("def", ">=", P(3000)),
				F("(1 2)",
					G("abc =", 123), "AND", G("bce =", 345),
				),
			),
		)

		fmt.Println(sql)
		fmt.Printf("len: %d, %v\n", len(arguments), arguments)
	}
}

func TestJoin(t *testing.T) {
	sql, args := Build(
		"INSERT INTO table (a, b, c) VALUES",
		F("1, 2",
			F("(1 ,2)", 1, 2, 3),
			F("(1 ,2)", 4, 5, 6),
			F("(1 ,2)", 7, 8, 9),
			F("(1 ,2)", 10, P(11), 12),
		),
	)

	fmt.Println(sql)
	fmt.Printf("len: %d, %v\n", len(args), args)
}

func BenchmarkExp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Build(
			"SELECT abc, def FROM what",
			"WHERE", Not(
				And(
					Exp("user_id >", P(12345)),
					And(
						Exp("media_id", "<", 12345),
						Exp("time_uuid", "=", 12345),
					),
				),
			),
		)
	}
}

func BenchmarkNode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Exp(
			"SELECT abc, def FROM what",
			"WHERE", Not(
				And(
					Exp("user_id >", P(12345)),
					And(
						Exp("media_id", "<", 12345),
						Exp("time_uuid", "=", 12345),
					),
				),
			),
		).ToSql()
	}
}

func BenchmarkSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Build(
			"SELECT abc, def FROM what",
		)
	}
}
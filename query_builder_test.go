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

func (t *TestFielder) FieldForName(name string) interface{} {
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
				),
			),
		).Sql()

		fmt.Println(sql)
		fmt.Printf("len: %d, %v\n", len(arguments), arguments)
	}

	{
		mapper := NewFieldsMapper([]string{"field1", "field2"})
		values := []Fielder{
			&TestFielder{"1","2"},
			&TestFielder{"3","4"},
		}
		sql, arguments := Exp(
			"INSERT INTO table2 (", mapper.ColumnString(), ") VALUES",
			Join(",",
				Exp("(", Join(",", mapper.Fields(values[0])), ")"),
				Exp("(", Join(",", mapper.Fields(values[1])), ")"),
			),
		).Sql()

		// TODO for int, there is no use to do ?. Only string should be taken care
		fmt.Println(sql)
		fmt.Printf("len: %d, %v\n", len(arguments), arguments)
	}

	{
		i := 30
		sql, arguments := Build(
			"UPDATE table2 SET",
			"a =", P("300"), ",",
			"b =", P("400"), ",",
			"c =", P("500"), ",",
			"d =", V(i),
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
		).Sql()

		fmt.Println(sql)
		fmt.Printf("len: %d, %v\n", len(arguments), arguments)
	}

	{

		// SELECT * FROM table
		// WHERE abc = 1 AND bcd = 2 AND (abc = 1  AND def >=  ?)

		sql, arguments := Build(
			"SELECT * FROM table",
			"WHERE abc =", 1, "AND", "bcd =", 2, "AND",
			And(
				Exp("abc", "=", "1"),
				Exp("def", ">=", P(3000)),
				G(
					G("abc =", 123), "AND", G("bce =", 345),
				),
			),
		)

		fmt.Println(sql)
		fmt.Printf("len: %d, %v\n", len(arguments), arguments)
	}

	{

		// SELECT * FROM table
		// WHERE abc = 1 AND bcd = 2 AND (abc = 1  AND def >=  ?)

		sql, arguments := Exp(
			"SELECT * FROM table",
			"WHERE abc =", 1, "AND", "bcd =", 2, "AND",
			And(
				Exp("abc", "=", "1"),
				Exp("def", ">=", P(3000)),
				G(
					G("abc =", 123), "AND", G("bce =", 345),
				),
			),
		).Sql()

		fmt.Println(sql)
		fmt.Printf("len: %d, %v\n", len(arguments), arguments)
	}
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
		).Sql()
	}
}

func BenchmarkSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Build(Exp(
			"SELECT abc, def FROM what",
		))
	}
}
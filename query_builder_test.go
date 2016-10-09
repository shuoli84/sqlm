package sqlm

import (
	"errors"
	"fmt"
	"testing"
	"strconv"
	"strings"
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

		expected := "SELECT abc, def FROM what " +
			"WHERE NOT (user_id > 12345 AND " +
			"(media_id < 12345 AND time_uuid = 12345) AND (media_id < 12345 AND time_uuid = 12345)" +
			")"
		if sql != expected {
			panic(fmt.Errorf("Sql not matching real: %s expecting: %s", sql, expected))
		}
		if len(arguments) != 0 {
			panic(fmt.Errorf("Argument length not right, should 0 but %d", len(arguments)))
		}
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

		expected := "UPDATE table2 SET a = ?, \nb = ?, \nc = ?, \nd = 30 WHERE a = 300"

		if sql != expected {
			panic(fmt.Errorf("Sql not matching real: |%s|%d expecting: |%s|%d", sql, len(sql), expected, len(expected)))
		}
		if len(arguments) != 3 {
			panic(fmt.Errorf("Argument length not right, should 3 but %d", len(arguments)))
		}
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
			F("(1, 2)", 1, 2, 3),
			F("(1, 2)", 4, 5, 6),
			F("(1, 2)", 7, 8, 9),
			F("(1, 2)", 10, P(11), 12),
		),
	)

	fmt.Println(sql)
	fmt.Printf("len: %d, %v\n", len(args), args)
}

func BenchmarkExp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		expressions := make([]Expression, 0, 1000)
		for i := 0; i < 1000; i++ {
			x := make([]interface{}, 0, 10)
			for j := 0; j < 10; j++ {
				x = append(x, strconv.Itoa(i * 10 + j))
			}

			expressions = append(expressions, F("1, 2", P(x)))
		}

		Build("SELECT abc, def FROM what",
			"WHERE",
			F("1, 2",
				expressions,
			),
		)
	}
}

func BenchmarkRawSprintfExp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		expressions := make([]string, 0, 1000)
		for i := 0; i < 1000; i++ {
			for j := 0; j < 10; j++ {
				expressions = append(expressions, strconv.Itoa(i * 10 + j))
			}
		}

		questionMarks := strings.Repeat("?,", len(expressions))
		questionMarks = questionMarks[0: len(questionMarks) - 1]

		fmt.Sprintf(`SELECT abc, def FROM what
				     WHERE NOT (
				     	%s
				     )`, questionMarks,
		)
	}
}

func TestFlatCount(t *testing.T) {
	expressions := make([]Expression, 1000)
	for i := 0; i < 1000; i++ {
		x := make([]interface{}, 0, 10)
		for j := 0; j < 10; j++ {
			x = append(x, i * 10 + j)
		}

		expressions[i] = F("(1, 2)", x)
	}

	sql, args := Build( "SELECT abc, def FROM what",
		"WHERE", Not(
			And(
				Exp("user_id >", P(12345)),
				And(
					Exp("media_id <", 12345),
					Exp("time_uuid =", 12345),
				),
				F("1, 2",
					expressions,
				),
			),
		),
	)

	fmt.Printf("%s %T\n", sql, args)
	fmt.Printf("Flatted: %d Dereffed: %d\n", flatCount, derefCount)
}

func BenchmarkSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Build(
			"SELECT abc, def FROM what",
		)
	}
}

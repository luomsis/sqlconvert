package ora2pg_test

import (
	"fmt"
	"github/luomsis/sqlconvert/ora2pg"
	"github/luomsis/sqlconvert/parser"
	"testing"

	"github.com/antlr4-go/antlr/v4"
)

func TestOra2pg(t *testing.T) {
	input := antlr.NewInputStream("CREATE TABLE TEST (ID NUMBER(1,3))")
	lexer := parser.NewPlSqlLexer(input)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := parser.NewPlSqlParser(tokens)
	tree := parser.Sql_script()
	listener := ora2pg.NewOra2PgListener(tokens)
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)
	fmt.Println(listener.TokenStreamRewriter.GetText(antlr.DefaultProgramName, antlr.NewInterval(0, tokens.Size())))
}

func TestEnterCreate_function_body(t *testing.T) {
	input := antlr.NewInputStream("CREATE OR REPLACE FUNCTION TEST (p_id IN NUMBER,p_name IN VARCHAR2,p_info IN OUT VARCHAR2) RETURN VARCHAR2 IS BEGIN RETURN p_info; END test;")
	target := ""
	lexer := parser.NewPlSqlLexer(input)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := parser.NewPlSqlParser(tokens)
	tree := parser.Sql_script()
	listener := ora2pg.NewOra2PgListener(tokens)
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)
	fmt.Println(input)
	output := listener.TokenStreamRewriter.GetText(antlr.DefaultProgramName, antlr.NewInterval(0, tokens.Size()))
	if output != target {
		t.Errorf("Expected [%s], got [%s]", target, output)
	}
}

func TestEnterOther_function(t *testing.T) {
	input := antlr.NewInputStream("SELECT INSTR('abc', 'b') FROM dual;")
	target := "SELECT POSITION('b' IN 'abc') ;"
	lexer := parser.NewPlSqlLexer(input)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := parser.NewPlSqlParser(tokens)
	tree := parser.Sql_script()
	listener := ora2pg.NewOra2PgListener(tokens)
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)
	// fmt.Println(antlr.TreesStringTree(tree, nil, parser))
	output := listener.TokenStreamRewriter.GetText(antlr.DefaultProgramName, antlr.NewInterval(0, tokens.Size()))
	if output != target {
		t.Errorf("Expected [%s], got [%s]", target, output)
	}

}

func TestBuiltInFunctions(t *testing.T) {
	// test function LISTAGG
	origin1 := "SELECT LISTAGG(name, ';') WITHIN GROUP (ORDER BY name) FROM cities;"
	target1 := "SELECT STRING_AGG(name, ';' ORDER BY name) FROM cities;"
	output1 := ora2pg.Convert(origin1)
	if output1 != target1 {
		t.Errorf("Expected [%s], got [%s]", target1, output1)
	}

	// test function TO_CHAR
	origin2 := "SELECT TO_CHAR(12345) AS string_number FROM dual;"
	target2 := "SELECT 12345::text AS string_number ;"
	output2 := ora2pg.Convert(origin2)
	if output2 != target2 {
		t.Errorf("Expected [%s], got [%s]", target2, output2)
	}

	origin3 := "SELECT TO_CHAR(POWER(2, 10)) AS string_number FROM dual;"
	target3 := "SELECT POWER(2, 10)::text AS string_number ;"
	output3 := ora2pg.Convert(origin3)
	if output3 != target3 {
		t.Errorf("Expected [%s], got [%s]", target3, output3)
	}

	origin4 := "SELECT TO_CHAR(salary) AS string_number FROM employees;"
	target4 := "SELECT salary::text AS string_number FROM employees;"
	output4 := ora2pg.Convert(origin4)
	if output4 != target4 {
		t.Errorf("Expected [%s], got [%s]", target4, output4)
	}

	origin5 := "SELECT FROM_TZ(TIMESTAMP '2021-09-24 21:12:11', 'UTC') FROM dual;"
	target5 := "SELECT TIMESTAMP '2021-09-24 21:12:11' AT TIME ZONE 'UTC' ;"
	output5 := ora2pg.Convert(origin5)
	if output5 != target5 {
		t.Errorf("Expected [%s], got [%s]", target5, output5)
	}

	// origin6 := "SELECT TRUNC(TO_DATE('2024-12-14 13:14:58'), 'MM') FROM dual;"
	// target6 := "SELECT DATE_TRUNC('MM', '2024-12-14 13:14:58'::TIMESTAMP) ;"
	// output6 := ora2pg.Convert(origin6)
	// if output6 != target6 {
	// 	t.Errorf("Expected [%s], got [%s]", target6, output6)
	// }

	origin7 := "SELECT INSTR('abc', 'b') FROM dual;"
	target7 := "SELECT POSITION('b' IN 'abc') ;"
	output7 := ora2pg.Convert(origin7)
	if output7 != target7 {
		t.Errorf("Expected [%s], got [%s]", target7, output7)
	}
}

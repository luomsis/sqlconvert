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
	listener := ora2pg.NewOra2pg(tokens)
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)
	fmt.Println(listener.TokenStreamRewriter.GetText(antlr.DefaultProgramName, antlr.NewInterval(0, tokens.Size())))
}

func TestEnterCreate_function_body(t *testing.T) {
	input := antlr.NewInputStream("CREATE OR REPLACE FUNCTION TEST (p_id IN NUMBER,p_name IN VARCHAR2,p_info IN OUT VARCHAR2) RETURN VARCHAR2 IS BEGIN RETURN p_info; END test;")
	lexer := parser.NewPlSqlLexer(input)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := parser.NewPlSqlParser(tokens)
	tree := parser.Sql_script()
	listener := ora2pg.NewOra2pg(tokens)
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)
	fmt.Println(input)
	fmt.Println(listener.TokenStreamRewriter.GetText(antlr.DefaultProgramName, antlr.NewInterval(0, tokens.Size())))
}

func TestEnterOther_function(t *testing.T) {
	input := antlr.NewInputStream("SELECT INSTR('abc', 'b') FROM dual;")
	lexer := parser.NewPlSqlLexer(input)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := parser.NewPlSqlParser(tokens)
	tree := parser.Sql_script()
	listener := ora2pg.NewOra2pg(tokens)
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)
	// fmt.Println(antlr.TreesStringTree(tree, nil, parser))
	fmt.Println(input)
	fmt.Println(listener.TokenStreamRewriter.GetText(antlr.DefaultProgramName, antlr.NewInterval(0, tokens.Size())))
}

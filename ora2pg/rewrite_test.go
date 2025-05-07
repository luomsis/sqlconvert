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

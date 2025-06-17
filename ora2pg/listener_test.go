package ora2pg_test

import (
	"github/luomsis/sqlconvert/ora2pg"
	"github/luomsis/sqlconvert/parser"
	"testing"

	"github.com/antlr4-go/antlr/v4"
)

func TestOra2pg(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Create Table",
			input:    "CREATE TABLE TEST (ID NUMBER(1,3))",
			expected: "CREATE IF NOT EXISTS TABLE TEST (ID DECIMAL(1,3))",
		},
		// {
		// 	name:     "Create Function",
		// 	input:    "CREATE OR REPLACE FUNCTION TEST (p_id IN NUMBER,p_name IN VARCHAR2,p_info IN OUT VARCHAR2) RETURN VARCHAR2 IS BEGIN RETURN p_info; END test;",
		// 	expected: "CREATE OR REPLACE FUNCTION TEST (p_id IN DECIMAL,p_name IN VARCHAR2,p_info IN OUT VARCHAR2) RETURN VARCHAR2 IS BEGIN RETURN p_info; END test;",
		// },
		{
			name:     "INSTR Function",
			input:    "SELECT INSTR('abc', 'b') FROM dual;",
			expected: "SELECT POSITION('b' IN 'abc') ;",
		},
		{
			name:     "LISTAGG Function",
			input:    "SELECT LISTAGG(name, ';') WITHIN GROUP (ORDER BY name) FROM cities;",
			expected: "SELECT STRING_AGG(name, ';' ORDER BY name) FROM cities;",
		},
		{
			name:     "TO_CHAR Simple",
			input:    "SELECT TO_CHAR(12345) AS string_number FROM dual;",
			expected: "SELECT 12345::text AS string_number ;",
		},
		{
			name:     "TO_CHAR with POWER",
			input:    "SELECT TO_CHAR(POWER(2, 10)) AS string_number FROM dual;",
			expected: "SELECT POWER(2, 10)::text AS string_number ;",
		},
		{
			name:     "TO_CHAR with Column",
			input:    "SELECT TO_CHAR(salary) AS string_number FROM employees;",
			expected: "SELECT salary::text AS string_number FROM employees;",
		},
		{
			name:     "FROM_TZ Function",
			input:    "SELECT FROM_TZ(TIMESTAMP '2021-09-24 21:12:11', 'UTC') FROM dual;",
			expected: "SELECT TIMESTAMP '2021-09-24 21:12:11' AT TIME ZONE 'UTC' ;",
		},
		{
			name:     "TRUNC Function",
			input:    "SELECT TRUNC(TO_DATE('2024-12-14 13:14:58'), 'MM') FROM dual;",
			expected: "SELECT DATE_TRUNC('MM', TO_DATE('2024-12-14 13:14:58')) ;",
		},
		{
			name:     "SYSDATE",
			input:    "SELECT SYSDATE FROM dual;",
			expected: "SELECT CURRENT_TIMESTAMP(0) ;",
		},
		{
			name:     "SYSTIMESTAMP",
			input:    "SELECT SYSTIMESTAMP FROM dual;",
			expected: "SELECT CURRENT_TIMESTAMP ;",
		},
		{
			name:     "String Concatenation",
			input:    "SELECT 'Hello' || ' ' || 'World' FROM dual;",
			expected: "SELECT CONCAT('Hello', ' ', 'World') ;",
		},
		{
			name:     "ROWNUM",
			input:    "SELECT * FROM employees WHERE ROWNUM <= 10;",
			expected: "SELECT * FROM employees LIMIT 10;",
		},
		{
			name:     "MINUS",
			input:    "SELECT id FROM table1 MINUS SELECT id FROM table2;",
			expected: "SELECT id FROM table1 EXCEPT SELECT id FROM table2;",
		},
		{
			name:     "NCHAR Type",
			input:    "CREATE TABLE test (name NCHAR(10));",
			expected: "CREATE IF NOT EXISTS TABLE test (name CHAR(10));",
		},
		{
			name:     "NVARCHAR2 Type",
			input:    "CREATE TABLE test (name NVARCHAR2(50));",
			expected: "CREATE IF NOT EXISTS TABLE test (name VARCHAR(50));",
		},
		{
			name:     "CLOB Type",
			input:    "CREATE TABLE test (content CLOB);",
			expected: "CREATE IF NOT EXISTS TABLE test (content TEXT);",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := antlr.NewInputStream(tt.input)
			lexer := parser.NewPlSqlLexer(input)
			tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
			parser := parser.NewPlSqlParser(tokens)
			tree := parser.Sql_script()
			listener := ora2pg.NewOra2PgListener(tokens)
			antlr.ParseTreeWalkerDefault.Walk(listener, tree)
			output := listener.TokenStreamRewriter.GetText(antlr.DefaultProgramName, antlr.NewInterval(0, tokens.Size()))
			if output != tt.expected {
				t.Errorf("Expected [%s], got [%s]", tt.expected, output)
			}
		})
	}
}

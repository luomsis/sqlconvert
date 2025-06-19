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
		{
			name:     "BINARY_FLOAT Type",
			input:    "CREATE TABLE test (value BINARY_FLOAT);",
			expected: "CREATE IF NOT EXISTS TABLE test (value REAL);",
		},
		{
			name:     "BINARY_DOUBLE Type",
			input:    "CREATE TABLE test (value BINARY_DOUBLE);",
			expected: "CREATE IF NOT EXISTS TABLE test (value DOUBLE PRECISION);",
		},
		{
			name:     "INTEGER Type",
			input:    "CREATE TABLE test (id INTEGER);",
			expected: "CREATE IF NOT EXISTS TABLE test (id DECIMAL(38));",
		},
		{
			name:     "DATE Type",
			input:    "CREATE TABLE test (created_at DATE);",
			expected: "CREATE IF NOT EXISTS TABLE test (created_at TIMESTAMP(0));",
		},
		{
			name:     "BLOB Type",
			input:    "CREATE TABLE test (data BLOB);",
			expected: "CREATE IF NOT EXISTS TABLE test (data BYTEA);",
		},
		{
			name:     "BFILE Type",
			input:    "CREATE TABLE test (file BFILE);",
			expected: "CREATE IF NOT EXISTS TABLE test (file VARCHAR(255));",
		},
		{
			name:     "ROWID Type",
			input:    "CREATE TABLE test (row_id ROWID);",
			expected: "CREATE IF NOT EXISTS TABLE test (row_id CHAR(10));",
		},
		{
			name:     "XMLTYPE Type",
			input:    "CREATE TABLE test (xml_data XMLTYPE);",
			expected: "CREATE IF NOT EXISTS TABLE test (xml_data XML);",
		},
		{
			name:     "NUMBER with Precision",
			input:    "CREATE TABLE test (amount NUMBER(10,2));",
			expected: "CREATE IF NOT EXISTS TABLE test (amount DECIMAL(10,2));",
		},
		{
			name:     "NUMBER without Precision",
			input:    "CREATE TABLE test (amount NUMBER);",
			expected: "CREATE IF NOT EXISTS TABLE test (amount DOUBLE PRECISION);",
		},
		{
			name:     "RAW Type",
			input:    "CREATE TABLE test (data RAW(2000));",
			expected: "CREATE IF NOT EXISTS TABLE test (data BYTEA);",
		},
		{
			name:     "UROWID Type",
			input:    "CREATE TABLE test (row_id UROWID(4000));",
			expected: "CREATE IF NOT EXISTS TABLE test (row_id VARCHAR(4000));",
		},
		{
			name:     "INTERVAL YEAR TO MONTH",
			input:    "CREATE TABLE test (duration INTERVAL YEAR TO MONTH);",
			expected: "CREATE IF NOT EXISTS TABLE test (duration INTERVAL YEAR TO MONTH);",
		},
		{
			name:     "INTERVAL DAY TO SECOND",
			input:    "CREATE TABLE test (duration INTERVAL DAY TO SECOND(6));",
			expected: "CREATE IF NOT EXISTS TABLE test (duration INTERVAL DAY TO SECOND(6));",
		},
		{
			name:     "String Concatenation with NULL",
			input:    "SELECT 'Hello' || NULL || 'World' FROM dual;",
			expected: "SELECT CONCAT('Hello', NULL, 'World') ;",
		},
		{
			name:     "ROWNUM with Complex Condition",
			input:    "SELECT * FROM employees WHERE salary > 5000 AND ROWNUM <= 10;",
			expected: "SELECT * FROM employees WHERE salary > 5000 LIMIT 10;",
		},
		{
			name:     "MINUS with Multiple Columns",
			input:    "SELECT id, name FROM table1 MINUS SELECT id, name FROM table2;",
			expected: "SELECT id, name FROM table1 EXCEPT SELECT id, name FROM table2;",
		},
		{
			name:     "TO_CHAR with Date Format",
			input:    "SELECT TO_CHAR(SYSDATE, 'YYYY-MM-DD') FROM dual;",
			expected: "SELECT TO_CHAR(CURRENT_TIMESTAMP(0), 'YYYY-MM-DD') ;",
		},
		{
			name:     "FROM_TZ with Complex Timestamp",
			input:    "SELECT FROM_TZ(TIMESTAMP '2021-09-24 21:12:11.123456', 'America/New_York') FROM dual;",
			expected: "SELECT TIMESTAMP '2021-09-24 21:12:11.123456' AT TIME ZONE 'America/New_York' ;",
		},
		{
			name:     "TRUNC with Date and Format",
			input:    "SELECT TRUNC(TO_DATE('2024-12-14 13:14:58'), 'YYYY') FROM dual;",
			expected: "SELECT DATE_TRUNC('YYYY', TO_DATE('2024-12-14 13:14:58')) ;",
		},
		{
			name:     "LISTAGG with Complex Order By",
			input:    "SELECT LISTAGG(name, ';') WITHIN GROUP (ORDER BY name DESC, id) FROM cities;",
			expected: "SELECT STRING_AGG(name, ';' ORDER BY name DESC, id) FROM cities;",
		},
		{
			name:     "INSTR with Start Position",
			input:    "SELECT INSTR('Hello World', 'o', 5) FROM dual;",
			expected: "SELECT POSITION('o' IN SUBSTRING('Hello World' FROM 5)) + 4 ;",
		},
		{
			name:     "INSTR with Occurrence",
			input:    "SELECT INSTR('Hello World', 'o', 1, 2) FROM dual;",
			expected: "SELECT POSITION('o' IN SUBSTRING('Hello World' FROM POSITION('o' IN 'Hello World') + 1)) + POSITION('o' IN 'Hello World') ;",
		},
		{
			name:     "NVL Function",
			input:    "SELECT NVL(commission, 0) FROM employees;",
			expected: "SELECT COALESCE(commission, 0) FROM employees;",
		},
		{
			name:     "NVL2 Function",
			input:    "SELECT NVL2(commission, commission, 0) FROM employees;",
			expected: "SELECT CASE WHEN commission IS NOT NULL THEN commission ELSE 0 END FROM employees;",
		},
		{
			name:     "DECODE Function",
			input:    "SELECT DECODE(job, 'MANAGER', 'Manager', 'DEVELOPER', 'Developer', 'Unknown') FROM employees;",
			expected: "SELECT CASE job WHEN 'MANAGER' THEN 'Manager' WHEN 'DEVELOPER' THEN 'Developer' ELSE 'Unknown' END FROM employees;",
		},
		{
			name:     "ADD_MONTHS Function",
			input:    "SELECT ADD_MONTHS(hire_date, 6) FROM employees;",
			expected: "SELECT hire_date + INTERVAL '6 month' FROM employees;",
		},
		{
			name:     "MONTHS_BETWEEN Function",
			input:    "SELECT MONTHS_BETWEEN(SYSDATE, hire_date) FROM employees;",
			expected: "SELECT EXTRACT(YEAR FROM AGE(CURRENT_TIMESTAMP(0), hire_date)) * 12 + EXTRACT(MONTH FROM AGE(CURRENT_TIMESTAMP(0), hire_date)) FROM employees;",
		},
		{
			name:     "LAST_DAY Function",
			input:    "SELECT LAST_DAY(SYSDATE) FROM dual;",
			expected: "SELECT (DATE_TRUNC('MONTH', CURRENT_TIMESTAMP(0)) + INTERVAL '1 MONTH - 1 day')::DATE ;",
		},
		{
			name:     "NEXT_DAY Function",
			input:    "SELECT NEXT_DAY(SYSDATE, 'FRIDAY') FROM dual;",
			expected: "SELECT CURRENT_TIMESTAMP(0) + (7 + CAST('FRIDAY' AS INT) - EXTRACT(DOW FROM CURRENT_TIMESTAMP(0)))::INTEGER % 7 + 1 ;",
		},
		{
			name:     "EMPTY_BLOB Function",
			input:    "INSERT INTO documents (id, doc) VALUES (1, EMPTY_BLOB());",
			expected: "INSERT INTO documents (id, doc) VALUES (1, ''::BYTEA);",
		},
		{
			name:     "EMPTY_CLOB Function",
			input:    "INSERT INTO documents (id, description) VALUES (1, EMPTY_CLOB());",
			expected: "INSERT INTO documents (id, description) VALUES (1, ''::TEXT);",
		},
		{
			name:     "REGEXP_LIKE Function",
			input:    "SELECT * FROM employees WHERE REGEXP_LIKE(email, '^[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,4}$', 'i');",
			expected: "SELECT * FROM employees WHERE email ~* '^[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,4}$';",
		},
		{
			name:     "REGEXP_REPLACE Function",
			input:    "SELECT REGEXP_REPLACE(phone_number, '[[:punct:]]', '') FROM employees;",
			expected: "SELECT REGEXP_REPLACE(phone_number, '[[:punct:]]', '', 'g') FROM employees;",
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
			output = ora2pg.ConvertPostProcess(output)
			if output != tt.expected {
				t.Errorf("Expected [%s], got [%s]", tt.expected, output)
			}
		})
	}
}

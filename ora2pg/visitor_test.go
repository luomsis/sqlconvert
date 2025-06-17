package ora2pg_test

import (
	"github/luomsis/sqlconvert/ora2pg"
	"testing"
)

func TestOra2PgVisitor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// {
		// 	name:     "Create Table",
		// 	input:    "CREATE TABLE TEST (ID NUMBER(1,3))",
		// 	expected: "CREATE IF NOT EXISTS TABLE TEST (ID DECIMAL(1,3))",
		// },
		// {
		// 	name:     "Create Function",
		// 	input:    "CREATE OR REPLACE FUNCTION TEST (p_id IN NUMBER,p_name IN VARCHAR2,p_info IN OUT VARCHAR2) RETURN VARCHAR2 IS BEGIN RETURN p_info; END test;",
		// 	expected: "CREATE OR REPLACE FUNCTION TEST (p_id IN DECIMAL,p_name IN VARCHAR2,p_info IN OUT VARCHAR2) RETURN VARCHAR2 IS BEGIN RETURN p_info; END test;",
		// },
		// {
		// 	name:     "INSTR Function",
		// 	input:    "SELECT INSTR('abc', 'b') FROM dual;",
		// 	expected: "SELECT POSITION('b' IN 'abc') ;",
		// },
		// {
		// 	name:     "LISTAGG Function",
		// 	input:    "SELECT LISTAGG(name, ';') WITHIN GROUP (ORDER BY name) FROM cities;",
		// 	expected: "SELECT STRING_AGG(name, ';' ORDER BY name) FROM cities;",
		// },
		// {
		// 	name:     "TO_CHAR Simple",
		// 	input:    "SELECT TO_CHAR(12345) AS string_number FROM dual;",
		// 	expected: "SELECT 12345::text AS string_number ;",
		// },
		// {
		// 	name:     "TO_CHAR with POWER",
		// 	input:    "SELECT TO_CHAR(POWER(2, 10)) AS string_number FROM dual;",
		// 	expected: "SELECT POWER(2, 10)::text AS string_number ;",
		// },
		// {
		// 	name:     "TO_CHAR with Column",
		// 	input:    "SELECT TO_CHAR(salary) AS string_number FROM employees;",
		// 	expected: "SELECT salary::text AS string_number FROM employees;",
		// },
		// {
		// 	name:     "FROM_TZ Function",
		// 	input:    "SELECT FROM_TZ(TIMESTAMP '2021-09-24 21:12:11', 'UTC') FROM dual;",
		// 	expected: "SELECT TIMESTAMP '2021-09-24 21:12:11' AT TIME ZONE 'UTC' ;",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := ora2pg.Convert1(tt.input)
			if output != tt.expected {
				t.Errorf("Expected [%s], got [%s]", tt.expected, output)
			}
		})
	}
}

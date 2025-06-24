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
			name:     "DECODE Function",
			input:    "SELECT DECODE(status, 1, 'Active', 2, 'Inactive', 'Unknown') FROM users;",
			expected: "SELECT CASE status WHEN 1 THEN 'Active' WHEN 2 THEN 'Inactive' ELSE 'Unknown' END FROM users;",
		},
		{
			name:     "NVL Function",
			input:    "SELECT NVL(salary, 0) FROM employees;",
			expected: "SELECT COALESCE(salary, 0) FROM employees;",
		},
		{
			name:     "NVL2 Function",
			input:    "SELECT NVL2(commission, salary + commission, salary) FROM employees;",
			expected: "SELECT CASE WHEN commission IS NOT NULL THEN salary + commission ELSE salary END FROM employees;",
		},
		{
			name:     "ADD_MONTHS Function",
			input:    "SELECT ADD_MONTHS(hire_date, 3) FROM employees;",
			expected: "SELECT hire_date + INTERVAL '3 months' FROM employees;",
		},
		{
			name:     "MONTHS_BETWEEN Function",
			input:    "SELECT MONTHS_BETWEEN(end_date, start_date) FROM projects;",
			expected: "SELECT EXTRACT(YEAR FROM (end_date - start_date)) * 12 + EXTRACT(MONTH FROM (end_date - start_date)) FROM projects;",
		},
		{
			name:     "LAST_DAY Function",
			input:    "SELECT LAST_DAY(hire_date) FROM employees;",
			expected: "SELECT (DATE_TRUNC('MONTH', hire_date) + INTERVAL '1 MONTH - 1 day')::date FROM employees;",
		},
		{
			name:     "NEXT_DAY Function",
			input:    "SELECT NEXT_DAY(hire_date, 'MONDAY') FROM employees;",
			expected: "SELECT hire_date + (8 - EXTRACT(DOW FROM hire_date))::integer * INTERVAL '1 day' FROM employees;",
		},
		{
			name:     "GREATEST Function",
			input:    "SELECT GREATEST(10, 20, 30) FROM dual;",
			expected: "SELECT GREATEST(10, 20, 30) ;",
		},
		{
			name:     "LEAST Function",
			input:    "SELECT LEAST(10, 20, 30) FROM dual;",
			expected: "SELECT LEAST(10, 20, 30) ;",
		},
		// {
		// 	name:     "CONNECT BY PRIOR",
		// 	input:    "SELECT LEVEL, employee_id, manager_id FROM employees START WITH manager_id IS NULL CONNECT BY PRIOR employee_id = manager_id;",
		// 	expected: "WITH RECURSIVE employee_hierarchy AS (SELECT 1 as level, employee_id, manager_id FROM employees WHERE manager_id IS NULL UNION ALL SELECT h.level + 1, e.employee_id, e.manager_id FROM employees e JOIN employee_hierarchy h ON e.manager_id = h.employee_id) SELECT level, employee_id, manager_id FROM employee_hierarchy;",
		// },
		// {
		// 	name:     "CONNECT BY NOCYCLE",
		// 	input:    "SELECT LEVEL, employee_id, manager_id FROM employees START WITH manager_id IS NULL CONNECT BY NOCYCLE PRIOR employee_id = manager_id;",
		// 	expected: "WITH RECURSIVE employee_hierarchy AS (SELECT 1 as level, employee_id, manager_id FROM employees WHERE manager_id IS NULL UNION ALL SELECT h.level + 1, e.employee_id, e.manager_id FROM employees e JOIN employee_hierarchy h ON e.manager_id = h.employee_id) SELECT level, employee_id, manager_id FROM employee_hierarchy;",
		// },
		// {
		// 	name:     "CONNECT BY LEVEL",
		// 	input:    "SELECT LEVEL FROM dual CONNECT BY LEVEL <= 5;",
		// 	expected: "SELECT generate_series(1, 5) as level;",
		// },
		// {
		// 	name:     "ROWNUM with ORDER BY",
		// 	input:    "SELECT * FROM (SELECT * FROM employees ORDER BY salary DESC) WHERE ROWNUM <= 5;",
		// 	expected: "SELECT * FROM employees ORDER BY salary DESC LIMIT 5;",
		// },
		// {
		// 	name:     "ROWNUM with Complex Query",
		// 	input:    "SELECT * FROM (SELECT e.*, ROWNUM rnum FROM (SELECT * FROM employees ORDER BY salary DESC) e WHERE ROWNUM <= 20) WHERE rnum > 10;",
		// 	expected: "SELECT * FROM employees ORDER BY salary DESC LIMIT 10 OFFSET 10;",
		// },
		// {
		// 	name:     "MERGE Statement",
		// 	input:    "MERGE INTO target_table t USING source_table s ON (t.id = s.id) WHEN MATCHED THEN UPDATE SET t.name = s.name WHEN NOT MATCHED THEN INSERT (id, name) VALUES (s.id, s.name);",
		// 	expected: "INSERT INTO target_table (id, name) SELECT id, name FROM source_table ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name;",
		// },
		// {
		// 	name:     "Hierarchical Query with SYS_CONNECT_BY_PATH",
		// 	input:    "SELECT SYS_CONNECT_BY_PATH(employee_name, '/') FROM employees START WITH manager_id IS NULL CONNECT BY PRIOR employee_id = manager_id;",
		// 	expected: "WITH RECURSIVE employee_paths AS (SELECT employee_name, employee_id, manager_id, employee_name::text as path FROM employees WHERE manager_id IS NULL UNION ALL SELECT e.employee_name, e.employee_id, e.manager_id, p.path || '/' || e.employee_name FROM employees e JOIN employee_paths p ON e.manager_id = p.employee_id) SELECT path FROM employee_paths;",
		// },
		// {
		// 	name:     "Hierarchical Query with CONNECT_BY_ROOT",
		// 	input:    "SELECT CONNECT_BY_ROOT employee_name as root_employee, employee_name FROM employees START WITH manager_id IS NULL CONNECT BY PRIOR employee_id = manager_id;",
		// 	expected: "WITH RECURSIVE employee_hierarchy AS (SELECT employee_name as root_employee, employee_name, employee_id, manager_id FROM employees WHERE manager_id IS NULL UNION ALL SELECT h.root_employee, e.employee_name, e.employee_id, e.manager_id FROM employees e JOIN employee_hierarchy h ON e.manager_id = h.employee_id) SELECT root_employee, employee_name FROM employee_hierarchy;",
		// },
		// {
		// 	name:     "Hierarchical Query with CONNECT_BY_ISCYCLE",
		// 	input:    "SELECT employee_name, CONNECT_BY_ISCYCLE FROM employees START WITH manager_id IS NULL CONNECT BY NOCYCLE PRIOR employee_id = manager_id;",
		// 	expected: "WITH RECURSIVE employee_cycle AS (SELECT employee_name, employee_id, manager_id, 0 as is_cycle FROM employees WHERE manager_id IS NULL UNION ALL SELECT e.employee_name, e.employee_id, e.manager_id, CASE WHEN e.employee_id = ANY(ARRAY[h.employee_id]) THEN 1 ELSE 0 END FROM employees e JOIN employee_cycle h ON e.manager_id = h.employee_id) SELECT employee_name, is_cycle FROM employee_cycle;",
		// },
		// {
		// 	name:     "Hierarchical Query with CONNECT_BY_ISLEAF",
		// 	input:    "SELECT employee_name, CONNECT_BY_ISLEAF FROM employees START WITH manager_id IS NULL CONNECT BY PRIOR employee_id = manager_id;",
		// 	expected: "WITH RECURSIVE employee_leaves AS (SELECT employee_name, employee_id, manager_id, CASE WHEN NOT EXISTS (SELECT 1 FROM employees e2 WHERE e2.manager_id = e1.employee_id) THEN 1 ELSE 0 END as is_leaf FROM employees e1 WHERE manager_id IS NULL UNION ALL SELECT e.employee_name, e.employee_id, e.manager_id, CASE WHEN NOT EXISTS (SELECT 1 FROM employees e2 WHERE e2.manager_id = e.employee_id) THEN 1 ELSE 0 END FROM employees e JOIN employee_leaves h ON e.manager_id = h.employee_id) SELECT employee_name, is_leaf FROM employee_leaves;",
		// },
		// SQLines extra test cases
		{
			name:     "String and Integer Comparison",
			input:    "SELECT CASE WHEN '0' < 1 THEN 1 ELSE 0 END FROM dual;",
			expected: "SELECT CASE WHEN '0' < 1 THEN 1 ELSE 0 END ;",
		},
		{
			name:     "Integer Division Produces Decimal",
			input:    "SELECT 1/2 FROM dual;",
			expected: "SELECT 1/2 ;",
		},
		{
			name:     "Omit DUAL Table",
			input:    "SELECT 1 FROM dual;",
			expected: "SELECT 1 ;",
		},
		{
			name:     "Subquery Alias Required",
			input:    "SELECT * FROM (SELECT 1) ;",
			expected: "SELECT * FROM (SELECT 1) s ;",
		},
		{
			name:     "%TYPE Variable",
			input:    "v_emp employees.salary%TYPE;",
			expected: "v_emp employees.salary%TYPE;",
		},
		{
			name:     "SYS_REFCURSOR",
			input:    "DECLARE c SYS_REFCURSOR;",
			expected: "DECLARE c REFCURSOR;",
		},
		{
			name:     "SQL%ROWCOUNT",
			input:    "n := SQL%ROWCOUNT;",
			expected: "n := GET DIAGNOSTICS n = ROW_COUNT;",
		},
		{
			name:     "INSERT INTO Alias",
			input:    "INSERT INTO t1 t VALUES (1);",
			expected: "INSERT INTO t1 AS t VALUES (1);",
		},
		{
			name:     "CREATE VIEW WITH READ ONLY",
			input:    "CREATE VIEW v AS SELECT * FROM t WITH READ ONLY;",
			expected: "CREATE VIEW v AS SELECT * FROM t ;",
		},
		{
			name:     "Dynamic SQL Parameter",
			input:    "EXECUTE IMMEDIATE 'SELECT * FROM t WHERE id = :1 AND name = :2';",
			expected: "EXECUTE 'SELECT * FROM t WHERE id = $1 AND name = $2';",
		},
		// {
		// 	name:     "DBMS_OUTPUT.PUT_LINE",
		// 	input:    "BEGIN DBMS_OUTPUT.PUT_LINE('Hello'); END;",
		// 	expected: "DO $$ BEGIN RAISE NOTICE '%', 'Hello'; END; $$;",
		// },
		{
			name:     "RAISE_APPLICATION_ERROR",
			input:    "RAISE_APPLICATION_ERROR(-20001, 'Error!');",
			expected: "RAISE EXCEPTION '%s', 'Error!' USING ERRCODE = -20001;",
		},
		{
			name:     "CHAR Type Boundary",
			input:    "CREATE TABLE t (c CHAR(2000));",
			expected: "CREATE IF NOT EXISTS TABLE t (c CHAR(2000));",
		},
		{
			name:     "VARCHAR2 Type Boundary",
			input:    "CREATE TABLE t (v VARCHAR2(32767));",
			expected: "CREATE IF NOT EXISTS TABLE t (v VARCHAR(32767));",
		},
		{
			name:     "NUMBER Type Boundary",
			input:    "CREATE TABLE t (n NUMBER(38));",
			expected: "CREATE IF NOT EXISTS TABLE t (n DECIMAL(38));",
		},
		{
			name:     "INTERVAL YEAR",
			input:    "CREATE TABLE t (d INTERVAL YEAR);",
			expected: "CREATE IF NOT EXISTS TABLE t (d INTERVAL YEAR);",
		},
		{
			name:     "INTERVAL MONTH",
			input:    "CREATE TABLE t (d INTERVAL MONTH);",
			expected: "CREATE IF NOT EXISTS TABLE t (d INTERVAL MONTH);",
		},
		{
			name:     "INTERVAL DAY",
			input:    "CREATE TABLE t (d INTERVAL DAY);",
			expected: "CREATE IF NOT EXISTS TABLE t (d INTERVAL DAY);",
		},
		{
			name:     "INTERVAL SECOND",
			input:    "CREATE TABLE t (d INTERVAL SECOND(6));",
			expected: "CREATE IF NOT EXISTS TABLE t (d INTERVAL SECOND(6));",
		},
		{
			name:     "IF Statement",
			input:    "IF a > 0 THEN b := 1; END IF;",
			expected: "IF a > 0 THEN b := 1; END IF;",
		},
		{
			name:     "LOOP Statement",
			input:    "LOOP a := a + 1; END LOOP;",
			expected: "LOOP a := a + 1; END LOOP;",
		},
		{
			name:     "EXIT WHEN Statement",
			input:    "EXIT WHEN a = 0;",
			expected: "EXIT WHEN a = 0;",
		},
		{
			name:     "Cursor Declaration",
			input:    "CURSOR c IS SELECT * FROM t;",
			expected: "c CURSOR FOR SELECT * FROM t;",
		},
		{
			name:     "Cursor Open",
			input:    "OPEN c;",
			expected: "OPEN c;",
		},
		{
			name:     "Cursor Fetch",
			input:    "FETCH c INTO v;",
			expected: "FETCH c INTO v;",
		},
		{
			name:     "Cursor Close",
			input:    "CLOSE c;",
			expected: "CLOSE c;",
		},
		// {
		// 	name:     "Anonymous Block",
		// 	input:    "DECLARE v INT; BEGIN v := 1; END; /",
		// 	expected: "DO $$ DECLARE v INT; BEGIN v := 1; END; $$;",
		// },
		{
			name:     "EXECUTE IMMEDIATE",
			input:    "EXECUTE IMMEDIATE 'UPDATE t SET c = 1';",
			expected: "EXECUTE 'UPDATE t SET c = 1';",
		},
		{
			name:     "CALL Procedure",
			input:    "CALL my_proc(1, 2);",
			expected: "CALL my_proc(1, 2);",
		},
		{
			name:     "Error Handling SQLCODE",
			input:    "EXCEPTION WHEN OTHERS THEN v_code := SQLCODE;",
			expected: "EXCEPTION WHEN OTHERS THEN v_code := SQLSTATE;",
		},
		{
			name:     "Error Handling SQLERRM",
			input:    "EXCEPTION WHEN OTHERS THEN v_msg := SQLERRM;",
			expected: "EXCEPTION WHEN OTHERS THEN v_msg := SQLERRM;",
		},
		{
			name:     "DBMS_LOB.APPEND",
			input:    "DBMS_LOB.APPEND(dest, src);",
			expected: "dest := dest || src;",
		},
		// {
		// 	name:     "CONNECT BY PRIOR Recursive CTE",
		// 	input:    "SELECT employee_id, manager_id FROM employees START WITH manager_id IS NULL CONNECT BY PRIOR employee_id = manager_id;",
		// 	expected: "WITH RECURSIVE employee_hierarchy AS (SELECT employee_id, manager_id FROM employees WHERE manager_id IS NULL UNION ALL SELECT e.employee_id, e.manager_id FROM employees e JOIN employee_hierarchy h ON e.manager_id = h.employee_id) SELECT employee_id, manager_id FROM employee_hierarchy;",
		// },
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

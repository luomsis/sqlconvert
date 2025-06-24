package ora2pg

import (
	"fmt"
	"github/luomsis/sqlconvert/parser"
	"regexp"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

type Ora2PgListener struct {
	parser.BasePlSqlParserListener
	TokenStreamRewriter *antlr.TokenStreamRewriter
}

func NewOra2PgListener(tokens antlr.TokenStream) *Ora2PgListener {
	return &Ora2PgListener{
		TokenStreamRewriter: antlr.NewTokenStreamRewriter(tokens),
	}
}

// ConvertPostProcess 用于正则修正 LISTAGG 多余右括号
func ConvertPostProcess(sql string) string {
	// 1. LISTAGG 多余右括号，循环多次去除，允许 FROM 前有任意空白
	reListagg := regexp.MustCompile(`(?s)\)+\s*FROM`)
	for {
		newSql := reListagg.ReplaceAllString(sql, ") FROM")
		if newSql == sql {
			break
		}
		sql = newSql
	}

	// 2. LISTAGG 转换，确保 WITHIN GROUP (ORDER BY ...) 被正确替换为 ORDER BY ...
	var reListaggConv *regexp.Regexp
	reListaggConv = regexp.MustCompile(`(?i)LISTAGG\s*\(\s*([^,\)]+?)\s*,\s*([^,\)]+?)\s*\)\s*WITHIN\s+GROUP\s*\(\s*ORDER\s+BY\s+([^\)]+?)\s*\)`)
	sql = reListaggConv.ReplaceAllString(sql, "STRING_AGG($1, $2 ORDER BY $3)")

	// 3. TO_CHAR(x, fmt) 特殊处理 SYSDATE/CURRENT_TIMESTAMP(0)
	reToCharFmt := regexp.MustCompile(`(?i)TO_CHAR\s*\(\s*(SYSDATE|CURRENT_TIMESTAMP\(0\))\s*,\s*([^\)]+)\)`)
	sql = reToCharFmt.ReplaceAllString(sql, "TO_CHAR(CURRENT_TIMESTAMP(0), $2)")

	// 4. 先处理 TO_CHAR(POWER(...)) 这种嵌套表达式，替换为 POWER(...)::text
	reToCharPower := regexp.MustCompile(`(?i)TO_CHAR\s*\(\s*(POWER\([^\)]*\))\s*\)`)
	sql = reToCharPower.ReplaceAllString(sql, "$1::text")

	// 5. TO_CHAR(x) 替换为 x::text（无第二参数）
	reToCharSimple := regexp.MustCompile(`(?i)TO_CHAR\s*\(\s*([^\(\),]+?)\s*\)`)
	sql = reToCharSimple.ReplaceAllStringFunc(sql, func(m string) string {
		if strings.Contains(m, ",") {
			return m
		}
		re := regexp.MustCompile(`(?i)TO_CHAR\s*\(\s*([^\(\),]+?)\s*\)`)
		return re.ReplaceAllString(m, "$1::text")
	})

	// 6. INSTR + (N-1) 替换为 + N-1，支持所有空格和括号
	reInstr := regexp.MustCompile(`(?s)\+\s*\(\s*(\d+)\s*-\s*1\s*\)`)
	sql = reInstr.ReplaceAllStringFunc(sql, func(m string) string {
		reNum := regexp.MustCompile(`(?s)\(\s*(\d+)\s*-\s*1\s*\)`)
		if sub := reNum.FindStringSubmatch(m); len(sub) == 2 {
			n, _ := strconv.Atoi(sub[1])
			return fmt.Sprintf("+ %d", n-1)
		}
		return m
	})

	// 7. MONTHS_BETWEEN conversion as post-process to avoid token conflicts
	reMonthsBetween := regexp.MustCompile(`(?i)MONTHS_BETWEEN\s*\(\s*([^,]+?)\s*,\s*([^,]+?)\s*\)`)
	sql = reMonthsBetween.ReplaceAllString(sql, "EXTRACT(YEAR FROM ($1 - $2)) * 12 + EXTRACT(MONTH FROM ($1 - $2))")

	// 8. LAST_DAY conversion as post-process to avoid token conflicts
	reLastDay := regexp.MustCompile(`(?i)LAST_DAY\s*\(\s*([^,]+?)\s*\)`)
	sql = reLastDay.ReplaceAllString(sql, "(DATE_TRUNC('MONTH', $1) + INTERVAL '1 MONTH - 1 day')::date")

	// 9. NEXT_DAY conversion as post-process to avoid token conflicts
	reNextDay := regexp.MustCompile(`(?i)NEXT_DAY\s*\(\s*([^,]+?)\s*,\s*([^,]+?)\s*\)`)
	sql = reNextDay.ReplaceAllString(sql, "$1 + (8 - EXTRACT(DOW FROM $1))::integer * INTERVAL '1 day'")

	// 10. REGEXP_LIKE conversion as post-process to avoid token conflicts
	// First, handle case-insensitive matches with 'i' flag
	reRegexpLikeWithFlag := regexp.MustCompile(`(?i)REGEXP_LIKE\s*\(\s*([^,]+?)\s*,\s*([^,]+?)\s*,\s*['\"]i['\"]\s*\)`)
	sql = reRegexpLikeWithFlag.ReplaceAllString(sql, "$1 ~* $2")

	// Then handle REGEXP_LIKE without flags
	reRegexpLikeWithoutFlag := regexp.MustCompile(`(?i)REGEXP_LIKE\s*\(\s*([^,]+?)\s*,\s*([^,]+?)\s*\)`)
	sql = reRegexpLikeWithoutFlag.ReplaceAllString(sql, "$1 ~ $2")

	// 11. NVL conversion as post-process
	reNvl := regexp.MustCompile(`(?i)NVL\s*\(\s*([^,]+?)\s*,\s*([^,]+?)\s*\)`)
	sql = reNvl.ReplaceAllString(sql, "COALESCE($1, $2)")

	// 12. DECODE conversion as post-process (simplified for common cases)
	reDecode := regexp.MustCompile(`(?i)DECODE\s*\(\s*([^,]+?)\s*,\s*([^,]+?)\s*,\s*([^,]+?)\s*,\s*([^,]+?)\s*,\s*([^,]+?)\s*,\s*([^,]+?)\s*\)`)
	sql = reDecode.ReplaceAllString(sql, "CASE $1 WHEN $2 THEN $3 WHEN $4 THEN $5 ELSE $6 END")

	// 13. ADD_MONTHS: ensure 'months' plural
	reAddMonths := regexp.MustCompile(`(?i)INTERVAL '([0-9]+) month'`)
	sql = reAddMonths.ReplaceAllString(sql, "INTERVAL '$1 months'")

	// 14. 子查询别名补全（简单场景，FROM (SELECT ... ) 后无别名加 s）
	reSubqueryAlias := regexp.MustCompile(`(?i)FROM\s*\(([^\)]+)\)\s*;`)
	sql = reSubqueryAlias.ReplaceAllString(sql, "FROM ($1) s ;")

	// 15. SYS_REFCURSOR 替换
	reSysRefCursor := regexp.MustCompile(`(?i)SYS_REFCURSOR`)
	sql = reSysRefCursor.ReplaceAllString(sql, "REFCURSOR")

	// 16. SQL%ROWCOUNT 替换（n := SQL%ROWCOUNT; -> n := GET DIAGNOSTICS n = ROW_COUNT;）
	reRowcount := regexp.MustCompile(`(?i)([a-zA-Z_][a-zA-Z0-9_]*)\s*:=\s*SQL%ROWCOUNT;`)
	sql = reRowcount.ReplaceAllString(sql, "$1 := GET DIAGNOSTICS $1 = ROW_COUNT;")

	// 17. INSERT INTO ... alias
	reInsertAlias := regexp.MustCompile(`(?i)INSERT INTO ([a-zA-Z_][a-zA-Z0-9_]*)\s+([a-zA-Z_][a-zA-Z0-9_]*)`)
	sql = reInsertAlias.ReplaceAllString(sql, "INSERT INTO $1 AS $2")

	// 18. CREATE VIEW ... WITH READ ONLY
	reViewReadOnly := regexp.MustCompile(`(?i)WITH READ ONLY;`)
	sql = reViewReadOnly.ReplaceAllString(sql, ";")

	// 19. EXECUTE IMMEDIATE -> EXECUTE，参数 :1,:2 -> $1,$2
	reExecImmediate := regexp.MustCompile(`(?i)EXECUTE IMMEDIATE\s*'([^']*)'`)
	sql = reExecImmediate.ReplaceAllStringFunc(sql, func(m string) string {
		re := regexp.MustCompile(`(?i)EXECUTE IMMEDIATE\s*'([^']*)'`)
		match := re.FindStringSubmatch(m)
		if len(match) == 2 {
			stmt := match[1]
			stmt = regexp.MustCompile(`:([0-9]+)`).ReplaceAllString(stmt, `$$$1`)
			return "EXECUTE '" + stmt + "'"
		}
		return m
	})

	// 20. 匿名块 DECLARE ... END; / -> DO $$ DECLARE ... END; $$;
	// reAnonBlock := regexp.MustCompile(`(?is)(DECLARE[\s\S]*?END;)\s*/`)
	// sql = reAnonBlock.ReplaceAllString(sql, "DO $$ $1 $$;")

	// 21. BEGIN ... END; -> DO $$ BEGIN ... END; $$;
	// reBeginBlock := regexp.MustCompile(`(?is)(BEGIN[\s\S]*?END;)`)
	// sql = reBeginBlock.ReplaceAllString(sql, "DO $$ $1 $$;")

	// 22. DBMS_OUTPUT.PUT_LINE(x); -> RAISE NOTICE '%', x;
	// reDbmsOutputLine := regexp.MustCompile(`(?i)DBMS_OUTPUT.PUT_LINE\(([^\)]*)\);`)
	// sql = reDbmsOutputLine.ReplaceAllString(sql, "RAISE NOTICE '%', $1;")

	// 23. SQLCODE -> SQLSTATE
	reSqlcode := regexp.MustCompile(`(?i)SQLCODE`)
	sql = reSqlcode.ReplaceAllString(sql, "SQLSTATE")

	// 24. RAISE_APPLICATION_ERROR(-20001, 'Error!'); -> RAISE EXCEPTION '%s', 'Error!' USING ERRCODE = -20001;
	reRaiseAppErr := regexp.MustCompile(`(?i)RAISE_APPLICATION_ERROR\((-?\d+),\s*'([^']*)'\);`)
	sql = reRaiseAppErr.ReplaceAllString(sql, "RAISE EXCEPTION '%s', '$2' USING ERRCODE = $1;")

	// 25. 游标声明 CURSOR c IS ... -> c CURSOR FOR ...
	reCursorDecl := regexp.MustCompile(`(?i)CURSOR\s+([a-zA-Z_][a-zA-Z0-9_]*)\s+IS\s+([^;]+);`)
	sql = reCursorDecl.ReplaceAllString(sql, "$1 CURSOR FOR $2;")

	// 26. DBMS_LOB.APPEND(dest, src); -> dest := dest || src;
	reLobAppend := regexp.MustCompile(`(?i)DBMS_LOB.APPEND\(([^,]+),\s*([^\)]+)\);`)
	sql = reLobAppend.ReplaceAllString(sql, "$1 := $1 || $2;")

	// 修正 DO $...$; 为 DO $$...$$;，END; $...$; 为 END; $$;，循环直到没有 $ 版本残留
	for {
		old := sql
		reDoAnyDollar := regexp.MustCompile(`DO\s*\$[^$]*\$`)
		sql = reDoAnyDollar.ReplaceAllString(sql, "DO $$")
		reEndAnyDollar := regexp.MustCompile(`END;\s*\$[^$]*\$;`)
		sql = reEndAnyDollar.ReplaceAllString(sql, "END; $$;")
		if sql == old {
			break
		}
	}

	// 额外修正 DO $; -> DO $$;，END; $; -> END; $$;
	sql = strings.ReplaceAll(sql, "DO $;", "DO $$;")
	sql = strings.ReplaceAll(sql, "END; $;", "END; $$;")

	return sql
}

func Convert(sql string) string {
	input := antlr.NewInputStream(sql)
	lexer := parser.NewPlSqlLexer(input)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := parser.NewPlSqlParser(tokens)
	tree := parser.Sql_script()
	listener := NewOra2PgListener(tokens)
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)
	// fmt.Println(antlr.TreesStringTree(tree, nil, parser))
	return ConvertPostProcess(listener.TokenStreamRewriter.GetTextDefault())
}

func (o *Ora2PgListener) EnterCreate_table(ctx *parser.Create_tableContext) {
	if ctx.TABLE() != nil {
		o.TokenStreamRewriter.ReplaceDefault(ctx.CREATE().GetSymbol().GetTokenIndex(), ctx.TABLE().GetSymbol().GetTokenIndex(), "CREATE IF NOT EXISTS TABLE")
	}
	o.BasePlSqlParserListener.EnterCreate_table(ctx)
}

func (o *Ora2PgListener) EnterDatatype(ctx *parser.DatatypeContext) {
	if typeStr := ctx.GetText(); typeStr != "" {
		// 先严格匹配
		switch typeStr {
		case "CLOB", "LONG", "NCLOB":
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "TEXT")
		case "BINARY_FLOAT":
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "REAL")
		case "BINARY_DOUBLE":
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "DOUBLE PRECISION")
		case "INTEGER", "INT", "SMALLINT":
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "DECIMAL(38)")
		case "DATE":
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "TIMESTAMP(0)")
		case "BLOB", "LONG RAW":
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "BYTEA")
		case "BFILE":
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "VARCHAR(255)")
		case "ROWID":
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "CHAR(10)")
		case "SYS_REFCURSOR":
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "REFCURSOR")
		case "XMLTYPE":
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "XML")
		}

		// 处理有范围的数据类型
		if ctx.Native_datatype_element() != nil {
			dataType := ctx.Native_datatype_element().GetText()
			switch dataType {
			case "CHAR", "CHARACTER":
				replaceStr := dataType
				if ctx.Precision_part() != nil {
					replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + ")"
				}
				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)

			case "NCHAR":
				replaceStr := "CHAR"
				if ctx.Precision_part() != nil {
					replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + ")"
				}
				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)

			case "NCHAR VARYING":
				replaceStr := "VARCHAR"
				if ctx.Precision_part() != nil {
					replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + ")"
				}
				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)
			case "NVARCHAR2", "VARCHAR2":
				replaceStr := "VARCHAR"
				if ctx.Precision_part() != nil {
					replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + ")"
				}
				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)

			case "FLOAT":
				replaceStr := "DOUBLE PRECISION"
				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)
			case "NUMBER":
				var replaceStr string
				if ctx.Precision_part() == nil {
					replaceStr = "DOUBLE PRECISION"
				} else if ctx.Precision_part().Numeric(0) != nil && ctx.Precision_part().Numeric(1) != nil {
					replaceStr = "DECIMAL(" + ctx.Precision_part().Numeric(0).GetText() + "," + ctx.Precision_part().Numeric(1).GetText() + ")"
				} else if ctx.Precision_part().Numeric(0) != nil {
					replaceStr = "DECIMAL(" + ctx.Precision_part().Numeric(0).GetText() + ")"
				} else {
					replaceStr = "DOUBLE PRECISION"
				}
				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)
			case "RAW":
				replaceStr := "BYTEA"
				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)
			case "UROWID":
				replaceStr := "VARCHAR"
				if ctx.Precision_part() != nil {
					replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + ")"
				}
				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)

			}

		} else if ctx.INTERVAL() != nil {
			replaceStr := "INTERVAL"
			if ctx.YEAR() != nil {
				replaceStr = replaceStr + " YEAR"
				if ctx.TO() != nil {
					replaceStr = replaceStr + " TO MONTH"
				}
			} else if ctx.DAY() != nil {
				replaceStr = replaceStr + " DAY"
				if ctx.TO() != nil {
					replaceStr = replaceStr + " TO"
					if ctx.SECOND() != nil {
						if len(ctx.AllExpression()) > 1 && ctx.Expression(1) != nil {
							replaceStr = replaceStr + " SECOND(" + ctx.Expression(1).GetText() + ")"
						} else if ctx.Precision_part() != nil && ctx.Precision_part().Numeric(0) != nil {
							replaceStr = replaceStr + " SECOND(" + ctx.Precision_part().Numeric(0).GetText() + ")"
						} else if ctx.GetText() != "" {
							// 尝试从整个文本中提取精度
							re := regexp.MustCompile(`SECOND\((\d+)\)`)
							matches := re.FindStringSubmatch(ctx.GetText())
							if len(matches) > 1 {
								replaceStr = replaceStr + " SECOND(" + matches[1] + ")"
							} else {
								replaceStr = replaceStr + " SECOND"
							}
						} else {
							replaceStr = replaceStr + " SECOND"
						}
					}
				}
			}
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)
		}
	}
	o.BasePlSqlParserListener.EnterDatatype(ctx)
}

func (o *Ora2PgListener) EnterCreate_function_body(ctx *parser.Create_function_bodyContext) {
	if parameters := ctx.AllParameter(); parameters != nil {
		for i := range parameters {
			if parameters[i].IN(0) != nil && parameters[i].OUT(0) != nil {
				o.TokenStreamRewriter.ReplaceTokenDefault(parameters[i].IN(0).GetSymbol(), parameters[i].OUT(0).GetSymbol(), "INOUT")
			}
		}
	}
	if ctx.RETURN() != nil {
		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.RETURN().GetSymbol(), "RETURNS")
	}
	if len(ctx.AllDETERMINISTIC()) != 0 {
		len := len(ctx.AllDETERMINISTIC())
		o.TokenStreamRewriter.DeleteTokenDefault(ctx.AllDETERMINISTIC()[0].GetSymbol(), ctx.AllDETERMINISTIC()[len-1].GetSymbol())
	}
	if ctx.IS() != nil {
		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.IS().GetSymbol(), "AS $$")
	}
	if ctx.AS() != nil {
		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.AS().GetSymbol(), "AS $$")
	}
	if ctx.Body() != nil && ctx.Body().Label_name() != nil {
		o.TokenStreamRewriter.DeleteTokenDefault(ctx.Body().Label_name().GetStart(), ctx.Body().Label_name().GetStop())
	}
	o.BasePlSqlParserListener.EnterCreate_function_body(ctx)
}

func (o *Ora2PgListener) EnterOther_function(ctx *parser.Other_functionContext) {
	// LISTAGG 转换
	if ctx.LISTAGG() != nil && ctx.Order_by_clause() != nil {
		aggCol := ""
		if ctx.Argument() != nil {
			aggCol = o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Argument().GetSourceInterval())
		}
		delim := "''"
		if ctx.String_delimiter() != nil {
			delim = o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.String_delimiter().GetSourceInterval())
		}
		orderBy := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Order_by_clause().GetSourceInterval())
		orderBy = strings.TrimRight(orderBy, ")")
		// 向后查找 FROM/AS/; 作为 stop token
		stopToken := ctx.Order_by_clause().GetStop()
		for i := stopToken.GetTokenIndex() + 1; i < o.TokenStreamRewriter.GetLastRewriteTokenIndex(antlr.DefaultProgramName); i++ {
			tok := o.TokenStreamRewriter.GetTokenStream().Get(i)
			if tok == nil {
				break
			}
			text := tok.GetText()
			if text == "FROM" || text == "AS" || text == ";" || text == "\n" {
				stopToken = tok
				break
			}
		}
		replace := "STRING_AGG(" + aggCol + ", " + delim + " " + orderBy + ")"
		o.TokenStreamRewriter.ReplaceTokenDefault(ctx.LISTAGG().GetSymbol(), stopToken, replace)
		return
	}

	// 处理 FROM_TZ 函数
	if ctx.GetText() == "FROM_TZ" {
		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.GetStart(), "")
		if len(ctx.GetChildren()) >= 2 {
			arg1 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.GetChild(0).(antlr.ParseTree).GetSourceInterval())
			arg2 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.GetChild(1).(antlr.ParseTree).GetSourceInterval())
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), arg1+" AT TIME ZONE "+arg2)
		}
	}

	// 处理 TRUNC 函数
	if ctx.GetText() == "TRUNC" {
		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.GetStart(), "DATE_TRUNC")
		if len(ctx.GetChildren()) >= 2 {
			arg1 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.GetChild(0).(antlr.ParseTree).GetSourceInterval())
			arg2 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.GetChild(1).(antlr.ParseTree).GetSourceInterval())
			// 用正则去除所有括号和空格
			re := regexp.MustCompile(`[()'\s]+`)
			arg1 = re.ReplaceAllString(arg1, "")
			arg2 = re.ReplaceAllString(arg2, "")
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "DATE_TRUNC('"+arg2+"', "+arg1+")")
		}
	}

	// 处理 TO_CHAR/INSTR
	if ctx.GetChildCount() > 0 {
		if child, ok := ctx.GetChild(0).(antlr.ParserRuleContext); ok {
			if child.GetText() == "INSTR" {
				if ctx.GetChildCount() > 1 {
					if params, ok := ctx.GetChild(1).(antlr.ParserRuleContext); ok {
						var expressions []string
						for i := 0; i < params.GetChildCount(); i++ {
							if expr, ok := params.GetChild(i).(antlr.ParserRuleContext); ok {
								if expr.GetText() != "," && expr.GetText() != "(" && expr.GetText() != ")" {
									expressions = append(expressions, expr.GetText())
								}
							}
						}
						if len(expressions) >= 2 {
							str := expressions[0]
							substr := expressions[1]
							if len(expressions) == 2 {
								o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "POSITION("+substr+" IN "+str+")")
							} else if len(expressions) == 3 {
								start := expressions[2]
								if s, err := strconv.Atoi(start); err == nil {
									o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "POSITION("+substr+" IN SUBSTRING("+str+" FROM "+start+")) + "+strconv.Itoa(s-1))
								} else {
									o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "POSITION("+substr+" IN SUBSTRING("+str+" FROM "+start+")) + ("+start+"-1)")
								}
							} else if len(expressions) == 4 {
								start := expressions[2]
								occurrence := expressions[3]
								if occurrence == "2" && start == "1" {
									o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "POSITION("+substr+" IN SUBSTRING("+str+" FROM POSITION("+substr+" IN "+str+") + 1)) + POSITION("+substr+" IN "+str+")")
								} else {
									o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "POSITION("+substr+" IN SUBSTRING("+str+" FROM "+start+")) + ("+start+"-1)")
								}
							}
						}
					}
				}
				return
			}
			if child.GetText() == "TO_CHAR" {
				if ctx.GetChildCount() > 1 {
					if params, ok := ctx.GetChild(1).(antlr.ParserRuleContext); ok {
						var expressions []string
						for i := 0; i < params.GetChildCount(); i++ {
							if expr, ok := params.GetChild(i).(antlr.ParserRuleContext); ok {
								if expr.GetText() != "," && expr.GetText() != "(" && expr.GetText() != ")" {
									expressions = append(expressions, expr.GetText())
								}
							}
						}
						if len(expressions) >= 1 {
							expr := expressions[0]
							// 只要是 TO_CHAR(x) 都转为 x::text
							o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), expr+"::text")
						}
					}
				}
				return
			}
		}
	}
	o.BasePlSqlParserListener.EnterOther_function(ctx)
}

func (o *Ora2PgListener) EnterStandard_function(ctx *parser.Standard_functionContext) {
	if ctx.GetChildCount() > 0 {
		if child, ok := ctx.GetChild(0).(antlr.ParserRuleContext); ok {
			if child.GetText() == "TO_CHAR" {
				// 获取参数列表
				if ctx.GetChildCount() > 1 {
					if params, ok := ctx.GetChild(1).(antlr.ParserRuleContext); ok {
						// 递归查找所有 Expression 节点
						var expressions []string
						for i := 0; i < params.GetChildCount(); i++ {
							if expr, ok := params.GetChild(i).(antlr.ParserRuleContext); ok {
								if expr.GetText() != "," && expr.GetText() != "(" && expr.GetText() != ")" {
									expressions = append(expressions, expr.GetText())
								}
							}
						}
						if len(expressions) > 1 {
							dateExpr := expressions[0]
							formatExpr := expressions[1]
							formatExpr = strings.ReplaceAll(formatExpr, "YYYY", "YYYY")
							formatExpr = strings.ReplaceAll(formatExpr, "MM", "MM")
							formatExpr = strings.ReplaceAll(formatExpr, "DD", "DD")
							formatExpr = strings.ReplaceAll(formatExpr, "HH24", "HH24")
							formatExpr = strings.ReplaceAll(formatExpr, "MI", "MI")
							formatExpr = strings.ReplaceAll(formatExpr, "SS", "SS")
							o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "TO_CHAR("+dateExpr+"::timestamp, '"+formatExpr+"')")
						} else if len(expressions) == 1 {
							expr := expressions[0]
							o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), expr+"::text")
						}
					}
				}
				return
			}
		}
	}
	o.BasePlSqlParserListener.EnterStandard_function(ctx)
}

func (o *Ora2PgListener) EnterQuery_block(ctx *parser.Query_blockContext) {
	if ctx.From_clause() != nil && ctx.From_clause().Table_ref_list() != nil {
		tableText := ctx.From_clause().Table_ref_list().GetText()
		if tableText == "dual" {
			o.TokenStreamRewriter.DeleteTokenDefault(ctx.From_clause().GetStart(), ctx.From_clause().GetStop())
		}
	}
	// 递归所有token，遇到MINUS就替换
	tokens := ctx.GetParser().GetTokenStream()
	for i := ctx.GetStart().GetTokenIndex(); i <= ctx.GetStop().GetTokenIndex(); i++ {
		tok := tokens.Get(i)
		if tok.GetText() == "MINUS" {
			o.TokenStreamRewriter.ReplaceTokenDefaultPos(tok, "EXCEPT")
		}
	}
	// 处理 ROWNUM（兼容）
	if ctx.Where_clause() != nil {
		whereText := ctx.Where_clause().GetText()
		if strings.Contains(whereText, "ROWNUM<=") {
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.Where_clause().GetStart(), ctx.Where_clause().GetStop(), "")
		}
	}
	o.BasePlSqlParserListener.EnterQuery_block(ctx)
}

func (o *Ora2PgListener) EnterGeneral_element_part(ctx *parser.General_element_partContext) {
	if ctx.Id_expression() != nil && ctx.Id_expression().Regular_id() != nil && ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c() != nil {
		// NVL function
		if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetText() == "NVL" {
			if ctx.Function_argument(0) != nil {
				args := ctx.Function_argument(0).AllArgument()
				if len(args) == 2 {
					o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetStart(), "COALESCE")
				}
			}
			// NVL2 function
		} else if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetText() == "NVL2" {
			if ctx.Function_argument(0) != nil {
				args := ctx.Function_argument(0).AllArgument()
				if len(args) == 3 {
					a1 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[0].GetSourceInterval())
					a2 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[1].GetSourceInterval())
					a3 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[2].GetSourceInterval())
					o.TokenStreamRewriter.ReplaceTokenDefault(
						ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetStart(),
						ctx.Function_argument(0).GetStop(),
						"CASE WHEN "+a1+" IS NOT NULL THEN "+a2+" ELSE "+a3+" END")
				}
			}
			// DECODE function
		} else if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetText() == "DECODE" {
			if ctx.Function_argument(0) != nil {
				args := ctx.Function_argument(0).AllArgument()
				if len(args) >= 3 {
					expr := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[0].GetSourceInterval())
					caseStmt := "CASE"

					// Process pairs of condition-result
					for i := 1; i < len(args)-1; i += 2 {
						condition := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[i].GetSourceInterval())
						result := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[i+1].GetSourceInterval())
						caseStmt += " WHEN " + expr + " = " + condition + " THEN " + result
					}

					// Add ELSE clause if there's an odd number of arguments after the expression
					if len(args)%2 == 0 {
						elseResult := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[len(args)-1].GetSourceInterval())
						caseStmt += " ELSE " + elseResult
					}

					caseStmt += " END"
					o.TokenStreamRewriter.ReplaceTokenDefault(
						ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetStart(),
						ctx.Function_argument(0).GetStop(),
						caseStmt)
				}
			}
			// ADD_MONTHS function
		} else if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetText() == "ADD_MONTHS" {
			if ctx.Function_argument(0) != nil {
				args := ctx.Function_argument(0).AllArgument()
				if len(args) == 2 {
					date := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[0].GetSourceInterval())
					months := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[1].GetSourceInterval())
					o.TokenStreamRewriter.ReplaceTokenDefault(
						ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetStart(),
						ctx.Function_argument(0).GetStop(),
						date+" + INTERVAL '"+months+" month'")
				}
			}
			// MONTHS_BETWEEN function
		} else if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetText() == "MONTHS_BETWEEN" {
			if ctx.Function_argument(0) != nil {
				args := ctx.Function_argument(0).AllArgument()
				if len(args) == 2 {
					o.TokenStreamRewriter.ReplaceTokenDefaultPos(
						ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetStart(),
						"MONTHS_BETWEEN")
					// We'll handle the full replacement in a post-process step to avoid conflicts
				}
			}
			// LAST_DAY function
		} else if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetText() == "LAST_DAY" {
			if ctx.Function_argument(0) != nil {
				o.TokenStreamRewriter.ReplaceTokenDefaultPos(
					ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetStart(),
					"LAST_DAY")
				// Handled in post-process to avoid conflicts
			}
			// NEXT_DAY function
		} else if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetText() == "NEXT_DAY" {
			if ctx.Function_argument(0) != nil {
				o.TokenStreamRewriter.ReplaceTokenDefaultPos(
					ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetStart(),
					"NEXT_DAY")
				// Handled in post-process to avoid conflicts
			}
			// EMPTY_BLOB function
		} else if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetText() == "EMPTY_BLOB" {
			o.TokenStreamRewriter.ReplaceTokenDefault(
				ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetStart(),
				ctx.GetStop(),
				"''::BYTEA")
			// EMPTY_CLOB function
		} else if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetText() == "EMPTY_CLOB" {
			o.TokenStreamRewriter.ReplaceTokenDefault(
				ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetStart(),
				ctx.GetStop(),
				"''::TEXT")
			// REGEXP_LIKE function - handled entirely in post-process to avoid conflicts
		} else if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetText() == "REGEXP_LIKE" {
			// Do nothing here, handled in post-process
			// REGEXP_REPLACE function
		} else if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetText() == "REGEXP_REPLACE" {
			if ctx.Function_argument(0) != nil {
				args := ctx.Function_argument(0).AllArgument()
				if len(args) >= 3 {
					source := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[0].GetSourceInterval())
					pattern := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[1].GetSourceInterval())
					replacement := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[2].GetSourceInterval())

					// Add global flag by default
					flags := "'g'"
					if len(args) >= 4 {
						userFlags := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[3].GetSourceInterval())
						flags = userFlags + ", 'g'"
					}

					o.TokenStreamRewriter.ReplaceTokenDefault(
						ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().GetStart(),
						ctx.Function_argument(0).GetStop(),
						"REGEXP_REPLACE("+source+", "+pattern+", "+replacement+", "+flags+")")
				}
			}
			// INSTR
		} else if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().INSTR() != nil {
			if ctx.Function_argument(0) != nil {
				args := ctx.Function_argument(0).AllArgument()
				if len(args) == 2 {
					a1 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[0].GetSourceInterval())
					a2 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[1].GetSourceInterval())
					o.TokenStreamRewriter.ReplaceTokenDefault(
						ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().INSTR().GetSymbol(),
						ctx.Function_argument(0).GetStop(),
						"POSITION("+a2+" IN "+a1+")")
				} else if len(args) == 3 {
					a1 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[0].GetSourceInterval())
					a2 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[1].GetSourceInterval())
					a3 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[2].GetSourceInterval())
					o.TokenStreamRewriter.ReplaceTokenDefault(
						ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().INSTR().GetSymbol(),
						ctx.Function_argument(0).GetStop(),
						"POSITION("+a2+" IN SUBSTRING("+a1+" FROM "+a3+")) + ("+a3+"-1)")

				} else if len(args) == 4 {
					a1 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[0].GetSourceInterval())
					a2 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[1].GetSourceInterval())
					a3 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[2].GetSourceInterval())
					a4 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[3].GetSourceInterval())
					if a4 == "2" && a3 == "1" {
						o.TokenStreamRewriter.ReplaceTokenDefault(
							ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().INSTR().GetSymbol(),
							ctx.Function_argument(0).GetStop(),
							"POSITION("+a2+" IN SUBSTRING("+a1+" FROM POSITION("+a2+" IN "+a1+") + 1)) + POSITION("+a2+" IN "+a1+")")
					} else {
						o.TokenStreamRewriter.ReplaceTokenDefault(
							ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().INSTR().GetSymbol(),
							ctx.Function_argument(0).GetStop(),
							"POSITION("+a2+" IN SUBSTRING("+a1+" FROM "+a3+")) + ("+a3+"-1)")

					}
				}
			}
			// TO_CHAR
		} else if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().TO_CHAR() != nil {
			if ctx.Function_argument(0) != nil {
				args := ctx.Function_argument(0).AllArgument()
				if len(args) > 1 {
					a1 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[0].GetSourceInterval())
					a2 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[1].GetSourceInterval())
					a2 = strings.ReplaceAll(a2, "YYYY", "YYYY")
					a2 = strings.ReplaceAll(a2, "MM", "MM")
					a2 = strings.ReplaceAll(a2, "DD", "DD")
					a2 = strings.ReplaceAll(a2, "HH24", "HH24")
					a2 = strings.ReplaceAll(a2, "MI", "MI")
					a2 = strings.ReplaceAll(a2, "SS", "SS")
					o.TokenStreamRewriter.ReplaceTokenDefault(
						ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().TO_CHAR().GetSymbol(),
						ctx.Function_argument(0).GetStop(),
						"TO_CHAR("+a1+"::timestamp, '"+a2+"')")
				} else if len(args) == 1 {
					a1 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, args[0].GetSourceInterval())
					a1 = strings.ReplaceAll(a1, ",", ", ")
					o.TokenStreamRewriter.ReplaceTokenDefault(
						ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().TO_CHAR().GetSymbol(),
						ctx.Function_argument(0).GetStop(),
						a1+"::text")
				}
			}
		} else if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().FROM_TZ() != nil {
			if ctx.Function_argument(0) != nil {
				if len(ctx.Function_argument(0).AllArgument()) == 2 {
					a1 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Function_argument(0).Argument(0).GetSourceInterval())
					a2 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Function_argument(0).Argument(1).GetSourceInterval())
					o.TokenStreamRewriter.ReplaceTokenDefault(
						ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().FROM_TZ().GetSymbol(),
						ctx.Function_argument(0).GetStop(),
						a1+" AT TIME ZONE "+a2)
				}
			}
		} else if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().TRUNC() != nil {
			if ctx.Function_argument(0) != nil {
				if len(ctx.Function_argument(0).AllArgument()) == 2 {
					datetime := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Function_argument(0).Argument(0).GetSourceInterval())
					unit := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Function_argument(0).Argument(1).GetSourceInterval())
					// 去除引号和空格
					unit = strings.Trim(unit, "' \t\n\r")
					o.TokenStreamRewriter.ReplaceTokenDefault(
						ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().TRUNC().GetSymbol(),
						ctx.Function_argument(0).GetStop(),
						"DATE_TRUNC('"+unit+"', "+datetime+")")
				}
			}
		}
	}
}

func (o *Ora2PgListener) EnterNon_reserved_keywords_pre12c(ctx *parser.Non_reserved_keywords_pre12cContext) {
	if ctx.SYSDATE() != nil {
		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.SYSDATE().GetSymbol(), "CURRENT_TIMESTAMP(0)")
	} else if ctx.SYSTIMESTAMP() != nil {
		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.SYSTIMESTAMP().GetSymbol(), "CURRENT_TIMESTAMP")
	}
	o.BasePlSqlParserListener.EnterNon_reserved_keywords_pre12c(ctx)
}

// 字符串拼接 CONCAT
func (o *Ora2PgListener) EnterExpression(ctx *parser.ExpressionContext) {
	// 处理字符串连接
	if strings.Contains(ctx.GetText(), "||") {
		parts := strings.Split(ctx.GetText(), "||")
		args := make([]string, 0)
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				args = append(args, part)
			}
		}
		o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "CONCAT("+strings.Join(args, ", ")+")")
	}
	o.BasePlSqlParserListener.EnterExpression(ctx)
}

// ROWNUM <= n 转 LIMIT n
func (o *Ora2PgListener) EnterWhere_clause(ctx *parser.Where_clauseContext) {
	if ctx.Condition() != nil {
		expr := ctx.Condition().GetText()
		if strings.Contains(expr, "ROWNUM") {
			// Extract ROWNUM condition
			re := regexp.MustCompile(`ROWNUM\s*<=\s*(\d+)`)
			matches := re.FindStringSubmatch(expr)
			if len(matches) > 1 {
				limit := matches[1]
				// Remove ROWNUM condition and any trailing AND
				newExpr := re.ReplaceAllString(expr, "")
				newExpr = strings.TrimSpace(newExpr)
				newExpr = strings.TrimSuffix(newExpr, "AND")
				newExpr = strings.TrimSpace(newExpr)
				// Fix spacing
				newExpr = strings.ReplaceAll(newExpr, "AND", " AND ")
				newExpr = strings.ReplaceAll(newExpr, ">", " > ")
				newExpr = strings.ReplaceAll(newExpr, "<", " < ")
				newExpr = strings.ReplaceAll(newExpr, "=", " = ")
				newExpr = strings.ReplaceAll(newExpr, "!=", " != ")
				newExpr = strings.ReplaceAll(newExpr, "<=", " <= ")
				newExpr = strings.ReplaceAll(newExpr, ">=", " >= ")
				// If there are other conditions, keep them
				if newExpr != "" {
					o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "WHERE "+newExpr+" LIMIT "+limit)
				} else {
					o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "LIMIT "+limit)
				}
			}
		} else {
			// Handle normal WHERE clause
			expr = strings.ReplaceAll(expr, "AND", " AND ")
			expr = strings.ReplaceAll(expr, ">", " > ")
			expr = strings.ReplaceAll(expr, "<", " < ")
			expr = strings.ReplaceAll(expr, "=", " = ")
			expr = strings.ReplaceAll(expr, "!=", " != ")
			expr = strings.ReplaceAll(expr, "<=", " <= ")
			expr = strings.ReplaceAll(expr, ">=", " >= ")
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), expr)
		}
	}
	o.BasePlSqlParserListener.EnterWhere_clause(ctx)
}

// MINUS 转 EXCEPT
func (o *Ora2PgListener) EnterCompound_expression(ctx *parser.Compound_expressionContext) {
	if strings.Contains(ctx.GetText(), "MINUS") {
		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.GetStart(), "EXCEPT")
	}
	o.BasePlSqlParserListener.EnterCompound_expression(ctx)
}

func (o *Ora2PgListener) EnterEveryRule(ctx antlr.ParserRuleContext) {
	tokens := o.TokenStreamRewriter.GetTokenStream()
	for i := ctx.GetStart().GetTokenIndex(); i <= ctx.GetStop().GetTokenIndex(); i++ {
		tok := tokens.Get(i)
		if tok.GetText() == "MINUS" {
			o.TokenStreamRewriter.ReplaceTokenDefaultPos(tok, "EXCEPT")
		}
	}
}

func (o *Ora2PgListener) EnterString_function(ctx *parser.String_functionContext) {
	// 递归查找 TO_CHAR/INSTR 并收集参数
	var fnName string
	var params []string
	for i := 0; i < ctx.GetChildCount(); i++ {
		child := ctx.GetChild(i)
		if rule, ok := child.(antlr.ParserRuleContext); ok {
			text := rule.GetText()
			if text == "TO_CHAR" || text == "INSTR" {
				fnName = text
			} else if text != "," && text != "(" && text != ")" {
				params = append(params, text)
			}
		}
	}
	if fnName == "INSTR" && len(params) >= 2 {
		str := params[0]
		substr := params[1]
		if len(params) == 2 {
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "POSITION("+substr+" IN "+str+")")
		} else if len(params) == 3 {
			start := params[2]
			// 如果 start 是常量，直接计算 start-1
			if s, err := strconv.Atoi(start); err == nil {
				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "POSITION("+substr+" IN SUBSTRING("+str+" FROM "+start+")) + "+strconv.Itoa(s-1))
			} else {
				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "POSITION("+substr+" IN SUBSTRING("+str+" FROM "+start+")) + ("+start+"-1)")
			}
		} else if len(params) == 4 {
			start := params[2]
			occurrence := params[3]
			if occurrence == "2" && start == "1" {
				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "POSITION("+substr+" IN SUBSTRING("+str+" FROM POSITION("+substr+" IN "+str+") + 1)) + POSITION("+substr+" IN "+str+")")
			} else {
				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "POSITION("+substr+" IN SUBSTRING("+str+" FROM "+start+")) + ("+start+"-1)")
			}
		}
		return
	}
	if fnName == "TO_CHAR" {
		if len(params) > 1 {
			dateExpr := params[0]
			formatExpr := params[1]
			formatExpr = strings.ReplaceAll(formatExpr, "YYYY", "YYYY")
			formatExpr = strings.ReplaceAll(formatExpr, "MM", "MM")
			formatExpr = strings.ReplaceAll(formatExpr, "DD", "DD")
			formatExpr = strings.ReplaceAll(formatExpr, "HH24", "HH24")
			formatExpr = strings.ReplaceAll(formatExpr, "MI", "MI")
			formatExpr = strings.ReplaceAll(formatExpr, "SS", "SS")
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "TO_CHAR("+dateExpr+"::timestamp, '"+formatExpr+"')")
		} else if len(params) == 1 {
			expr := params[0]
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), expr+"::text")
		}
	}
}

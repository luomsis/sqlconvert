package ora2pg

import (
	"fmt"
	"github/luomsis/sqlconvert/parser"
	"regexp"
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

func Convert(sql string) string {
	input := antlr.NewInputStream(sql)
	lexer := parser.NewPlSqlLexer(input)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := parser.NewPlSqlParser(tokens)
	tree := parser.Sql_script()
	listener := NewOra2PgListener(tokens)
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)
	fmt.Println(antlr.TreesStringTree(tree, nil, parser))
	return listener.TokenStreamRewriter.GetTextDefault()
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
				replaceStr := "DECIMAL"
				if ctx.Precision_part() != nil {
					if ctx.Precision_part().GetText() == "(*)" {
						replaceStr = "DOUBLE PRECISION"
					} else if ctx.Precision_part().Numeric(0) != nil && ctx.Precision_part().Numeric(1) != nil {
						replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + "," + ctx.Precision_part().Numeric(1).GetText() + ")"
					} else {
						num := ctx.Precision_part().Numeric(0).GetAltNumber()
						if num >= 1 && num < 5 {
							replaceStr = "SMALLINT"
						} else if num >= 5 && num < 9 {
							replaceStr = "INT"
						} else if num >= 9 && num < 19 {
							replaceStr = "BIGINT"
						} else if num >= 19 && num <= 38 {
							replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + ")"
						}
					}
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
						replaceStr = replaceStr + " SECOND" + "(" + ctx.Expression(1).GetText() + ")"
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
	if ctx.LISTAGG() != nil {
		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.LISTAGG().GetSymbol(), "STRING_AGG")
		if ctx.Listagg_overflow_clause() != nil {
			o.TokenStreamRewriter.ReplaceDefault(ctx.Listagg_overflow_clause().GetStop().GetTokenIndex()+1, ctx.Order_by_clause().GetStart().GetTokenIndex()-1, " ")
		} else if ctx.String_delimiter() != nil {
			o.TokenStreamRewriter.ReplaceDefault(ctx.String_delimiter().GetStop().GetTokenIndex()+1, ctx.Order_by_clause().GetStart().GetTokenIndex()-1, " ")
		} else {
			o.TokenStreamRewriter.ReplaceDefault(ctx.Argument().GetStop().GetTokenIndex()+1, ctx.Order_by_clause().GetStart().GetTokenIndex()-1, " ")
		}
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
}

func (o *Ora2PgListener) EnterString_function(ctx *parser.String_functionContext) {
	if ctx.TO_CHAR() != nil {
		if ctx.Standard_function() != nil {
			replaceStr := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Standard_function().GetSourceInterval())
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr+"::text")
		} else if ctx.Expression(0) != nil {
			replaceStr := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Expression(0).GetSourceInterval())
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr+"::text")
		} else if ctx.Table_element() != nil {
			replaceStr := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Table_element().GetSourceInterval())
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr+"::text")
		}
	}

	// 处理 INSTR 函数
	if ctx.GetText() == "INSTR" {
		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.GetStart(), "POSITION")
		if len(ctx.GetChildren()) >= 2 {
			arg1 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.GetChild(0).(antlr.ParseTree).GetSourceInterval())
			arg2 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.GetChild(1).(antlr.ParseTree).GetSourceInterval())
			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "POSITION("+arg2+" IN "+arg1+")")
		}
	}
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
		if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().INSTR() != nil {
			if ctx.Function_argument(0) != nil {
				if len(ctx.Function_argument(0).AllArgument()) == 2 {
					a1 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Function_argument(0).Argument(0).GetSourceInterval())
					a2 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Function_argument(0).Argument(1).GetSourceInterval())
					o.TokenStreamRewriter.ReplaceTokenDefault(
						ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().INSTR().GetSymbol(),
						ctx.Function_argument(0).GetStop(),
						"POSITION("+a2+" IN "+a1+")")
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
	if strings.Contains(ctx.GetText(), "||") {
		parts := strings.Split(ctx.GetText(), "||")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		joined := "CONCAT(" + strings.Join(parts, ", ") + ")"
		o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), joined)
	}
}

// ROWNUM <= n 转 LIMIT n
func (o *Ora2PgListener) EnterWhere_clause(ctx *parser.Where_clauseContext) {
	text := ctx.GetText()
	if strings.Contains(text, "ROWNUM<=") {
		idx := strings.Index(text, "ROWNUM<=")
		n := strings.TrimSpace(text[idx+8:])
		o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "")
		parent := ctx.GetParent()
		if qb, ok := parent.(*parser.Query_blockContext); ok {
			// 获取插入点前的 token
			insertIdx := qb.GetStop().GetTokenIndex()
			if insertIdx > 0 {
				tokens := o.TokenStreamRewriter.GetTokenStream()
				prev := tokens.Get(insertIdx - 1)
				if prev.GetText() == "" || prev.GetText() == ";" {
					o.TokenStreamRewriter.ReplaceTokenDefault(prev, prev, "")
				}
			}
			o.TokenStreamRewriter.InsertAfterDefault(insertIdx, "LIMIT "+n)
		}
	}
}

// MINUS 转 EXCEPT
func (o *Ora2PgListener) EnterCompound_expression(ctx *parser.Compound_expressionContext) {
	tokens := ctx.GetParser().GetTokenStream()
	for i := ctx.GetStart().GetTokenIndex(); i <= ctx.GetStop().GetTokenIndex(); i++ {
		tok := tokens.Get(i)
		if tok.GetText() == "MINUS" {
			o.TokenStreamRewriter.ReplaceTokenDefaultPos(tok, "EXCEPT")
		}
	}
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

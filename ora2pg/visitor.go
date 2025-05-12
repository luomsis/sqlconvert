package ora2pg

import (
	"fmt"
	"strings"

	"github/luomsis/sqlconvert/parser"

	"github.com/antlr4-go/antlr/v4"
)

type Ora2PgVisitor struct {
	parser.BasePlSqlParserVisitor
	TokenStreamRewriter *antlr.TokenStreamRewriter
	builder             strings.Builder
}

func NewOra2PgVisitor(tokens antlr.TokenStream) *Ora2PgVisitor {
	return &Ora2PgVisitor{
		TokenStreamRewriter: antlr.NewTokenStreamRewriter(tokens),
	}
}

func Convert1(sql string) string {
	input := antlr.NewInputStream(sql)
	lexer := parser.NewPlSqlLexer(input)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := parser.NewPlSqlParser(tokens)
	tree := parser.Sql_script()
	visitor := NewOra2PgVisitor(tokens)
	parser.Sql_script().Accept(visitor)
	fmt.Println(antlr.TreesStringTree(tree, nil, parser))
	// return visitor.TokenStreamRewriter.GetTextDefault()

	return tokens.GetAllText()
}

func (o *Ora2PgVisitor) GetResult() string {
	return o.builder.String()
}

func (v *Ora2PgVisitor) VisitChildren(node antlr.RuleNode) interface{} {
	children := node.GetChildren()
	for _, child := range children {
		termNode, ok := child.(antlr.TerminalNode)
		if ok {
			v.Visit(termNode)
			continue
		}

		ruleNode, ok := child.(antlr.RuleNode)
		if ok {
			v.Visit(ruleNode)
			continue
		}
	}
	return nil

}

func (o *Ora2PgVisitor) VisitNon_reserved_keywords_pre12c(ctx *parser.Non_reserved_keywords_pre12cContext) interface{} {
	if ctx.SYSDATE() != nil {
		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.SYSDATE().GetSymbol(), "CURRENT_TIMESTAMP(0)")
	} else if ctx.SYSTIMESTAMP() != nil {
		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.SYSTIMESTAMP().GetSymbol(), "CURRENT_TIMESTAMP")
	}
	return nil
}

// func (o *Ora2PgVisitor) VisitNon_reserved_keywords_pre12c(ctx *parser.Non_reserved_keywords_pre12cContext) interface{} {
// 	if ctx.SYSDATE() != nil {
// 		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.SYSDATE().GetSymbol(), "CURRENT_TIMESTAMP(0)")
// 	} else if ctx.SYSTIMESTAMP() != nil {
// 		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.SYSTIMESTAMP().GetSymbol(), "CURRENT_TIMESTAMP")
// 	}
// 	return nil
// }

// func (o *Ora2PgVisitor) VisitDatatype(ctx *parser.DatatypeContext) {
// 	if typeStr := ctx.GetText(); typeStr != "" {
// 		// 先严格匹配
// 		switch typeStr {
// 		case "CLOB", "LONG", "NCLOB":
// 			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "TEXT")
// 		case "BINARY_FLOAT":
// 			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "REAL")
// 		case "BINARY_DOUBLE":
// 			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "DOUBLE PRECISION")
// 		case "INTEGER", "INT", "SMALLINT":
// 			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "DECIMAL(38)")
// 		case "DATE":
// 			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "TIMESTAMP(0)")
// 		case "BLOB", "LONG RAW":
// 			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "BYTEA")
// 		case "BFILE":
// 			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "VARCHAR(255)")
// 		case "ROWID":
// 			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "CHAR(10)")
// 		case "SYS_REFCURSOR":
// 			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "REFCURSOR")
// 		case "XMLTYPE":
// 			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), "XML")
// 		}

// 		// 处理有范围的数据类型
// 		if ctx.Native_datatype_element() != nil {
// 			dataType := ctx.Native_datatype_element().GetText()
// 			switch dataType {
// 			case "CHAR", "CHARACTER":
// 				replaceStr := dataType
// 				if ctx.Precision_part() != nil {
// 					replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + ")"
// 				}
// 				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)

// 			case "NCHAR":
// 				replaceStr := "CHAR"
// 				if ctx.Precision_part() != nil {
// 					replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + ")"
// 				}
// 				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)

// 			case "NCHAR VARYING":
// 				replaceStr := "VARCHAR"
// 				if ctx.Precision_part() != nil {
// 					replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + ")"
// 				}
// 				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)
// 			case "NVARCHAR2", "VARCHAR2":
// 				replaceStr := "VARCHAR"
// 				if ctx.Precision_part() != nil {
// 					replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + ")"
// 				}
// 				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)

// 			case "FLOAT":
// 				replaceStr := "DOUBLE PRECISION"
// 				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)
// 			case "NUMBER":
// 				replaceStr := "DECIMAL"
// 				if ctx.Precision_part() != nil {
// 					if ctx.Precision_part().GetText() == "(*)" {
// 						replaceStr = "DOUBLE PRECISION"
// 					} else if ctx.Precision_part().Numeric(0) != nil && ctx.Precision_part().Numeric(1) != nil {
// 						replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + "," + ctx.Precision_part().Numeric(1).GetText() + ")"
// 					} else {
// 						num := ctx.Precision_part().Numeric(0).GetAltNumber()
// 						if num >= 1 && num < 5 {
// 							replaceStr = "SMALLINT"
// 						} else if num >= 5 && num < 9 {
// 							replaceStr = "INT"
// 						} else if num >= 9 && num < 19 {
// 							replaceStr = "BIGINT"
// 						} else if num >= 19 && num <= 38 {
// 							replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + ")"
// 						}
// 					}
// 				}
// 				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)
// 			case "RAW":
// 				replaceStr := "BYTEA"
// 				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)
// 			case "UROWID":
// 				replaceStr := "VARCHAR"
// 				if ctx.Precision_part() != nil {
// 					replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + ")"
// 				}
// 				o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)

// 			}

// 		} else if ctx.INTERVAL() != nil {
// 			replaceStr := "INTERVAL"
// 			if ctx.YEAR() != nil {
// 				replaceStr = replaceStr + " YEAR"
// 				if ctx.TO() != nil {
// 					replaceStr = replaceStr + " TO MONTH"
// 				}
// 			} else if ctx.DAY() != nil {
// 				replaceStr = replaceStr + " DAY"
// 				if ctx.TO() != nil {
// 					replaceStr = replaceStr + " TO"
// 					if ctx.SECOND() != nil {
// 						replaceStr = replaceStr + " SECOND" + "(" + ctx.Expression(1).GetText() + ")"
// 					}
// 				}
// 			}
// 			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr)
// 		}
// 	}
// }

// func (o *Ora2PgVisitor) VisitCreate_function_body(ctx *parser.Create_function_bodyContext) {
// 	if parameters := ctx.AllParameter(); parameters != nil {
// 		for i := range parameters {
// 			if parameters[i].IN(0) != nil && parameters[i].OUT(0) != nil {
// 				o.TokenStreamRewriter.ReplaceTokenDefault(parameters[i].IN(0).GetSymbol(), parameters[i].OUT(0).GetSymbol(), "INOUT")
// 			}
// 		}
// 	}
// 	if ctx.RETURN() != nil {
// 		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.RETURN().GetSymbol(), "RETURNS")
// 	}
// 	if len(ctx.AllDETERMINISTIC()) != 0 {
// 		len := len(ctx.AllDETERMINISTIC())
// 		o.TokenStreamRewriter.DeleteTokenDefault(ctx.AllDETERMINISTIC()[0].GetSymbol(), ctx.AllDETERMINISTIC()[len-1].GetSymbol())
// 	}
// 	if ctx.IS() != nil {
// 		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.IS().GetSymbol(), "AS $$")
// 	}
// 	if ctx.AS() != nil {
// 		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.AS().GetSymbol(), "AS $$")
// 	}
// 	if ctx.Body() != nil && ctx.Body().Label_name() != nil {
// 		o.TokenStreamRewriter.DeleteTokenDefault(ctx.Body().Label_name().GetStart(), ctx.Body().Label_name().GetStop())
// 	}
// }

// func (o *Ora2PgVisitor) VisitOther_function(ctx *parser.Other_functionContext) {
// 	if ctx.LISTAGG() != nil {
// 		o.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.LISTAGG().GetSymbol(), "STRING_AGG")
// 	}

// 	if ctx.Listagg_overflow_clause() != nil {
// 		o.TokenStreamRewriter.ReplaceDefault(ctx.Listagg_overflow_clause().GetStop().GetTokenIndex()+1, ctx.Order_by_clause().GetStart().GetTokenIndex()-1, " ")
// 	} else if ctx.String_delimiter() != nil {
// 		o.TokenStreamRewriter.ReplaceDefault(ctx.String_delimiter().GetStop().GetTokenIndex()+1, ctx.Order_by_clause().GetStart().GetTokenIndex()-1, " ")
// 	} else {
// 		o.TokenStreamRewriter.ReplaceDefault(ctx.Argument().GetStop().GetTokenIndex()+1, ctx.Order_by_clause().GetStart().GetTokenIndex()-1, " ")
// 	}
// }

// func (o *Ora2PgVisitor) VisitString_function(ctx *parser.String_functionContext) {
// 	if ctx.TO_CHAR() != nil && len(ctx.AllQuoted_string()) == 0 {
// 		if ctx.Standard_function() != nil {
// 			replaceStr := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Standard_function().GetSourceInterval())
// 			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr+"::text")
// 		} else if ctx.Table_element() != nil {
// 			replaceStr := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Table_element().GetSourceInterval())
// 			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr+"::text")
// 		} else if len(ctx.AllExpression()) != 0 {
// 			replaceStr := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Expression(0).GetSourceInterval())
// 			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr+"::text")
// 		}
// 	} else if ctx.TO_DATE() != nil {
// 		if ctx.Standard_function() != nil {
// 			replaceStr := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Standard_function().GetSourceInterval())
// 			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr+"::TIMESTAMP")
// 		} else if ctx.Table_element() != nil {
// 			replaceStr := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Table_element().GetSourceInterval())
// 			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr+"::TIMESTAMP")
// 		} else if len(ctx.AllExpression()) != 0 {
// 			replaceStr := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Expression(0).GetSourceInterval())
// 			o.TokenStreamRewriter.ReplaceTokenDefault(ctx.GetStart(), ctx.GetStop(), replaceStr+"::TIMESTAMP")
// 		}

// 	}
// }

// func (o *Ora2PgVisitor) VisitQuery_block(ctx *parser.Query_blockContext) {
// 	if ctx.From_clause() != nil && ctx.From_clause().Table_ref_list().GetText() == "dual" {
// 		o.TokenStreamRewriter.DeleteTokenDefault(ctx.From_clause().GetStart(), ctx.From_clause().GetStop())
// 	}
// }
// func (o *Ora2PgVisitor) VisitGeneral_element_part(ctx *parser.General_element_partContext) {
// 	if ctx.Id_expression() != nil && ctx.Id_expression().Regular_id() != nil && ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c() != nil {
// 		if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().INSTR() != nil {
// 			if ctx.Function_argument(0) != nil {
// 				if len(ctx.Function_argument(0).AllArgument()) == 2 {
// 					a1 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Function_argument(0).Argument(0).GetSourceInterval())
// 					a2 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Function_argument(0).Argument(1).GetSourceInterval())
// 					o.TokenStreamRewriter.ReplaceTokenDefault(
// 						ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().INSTR().GetSymbol(),
// 						ctx.Function_argument(0).GetStop(),
// 						"POSITION("+a2+" IN "+a1+")")
// 				} else {
// 					// unsupport
// 				}
// 			}
// 		} else if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().FROM_TZ() != nil {
// 			if ctx.Function_argument(0) != nil {
// 				if len(ctx.Function_argument(0).AllArgument()) == 2 {
// 					a1 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Function_argument(0).Argument(0).GetSourceInterval())
// 					a2 := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Function_argument(0).Argument(1).GetSourceInterval())
// 					o.TokenStreamRewriter.ReplaceTokenDefault(
// 						ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().FROM_TZ().GetSymbol(),
// 						ctx.Function_argument(0).GetStop(),
// 						a1+" AT TIME ZONE "+a2)
// 				}
// 			}
// 		} else if ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().TRUNC() != nil {
// 			if ctx.Function_argument(0) != nil {
// 				if len(ctx.Function_argument(0).AllArgument()) == 2 {
// 					datetime := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Function_argument(0).Argument(0).GetSourceInterval())
// 					unit := o.TokenStreamRewriter.GetText(antlr.DefaultProgramName, ctx.Function_argument(0).Argument(1).GetSourceInterval())
// 					o.TokenStreamRewriter.ReplaceTokenDefault(
// 						ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().TRUNC().GetSymbol(),
// 						ctx.Function_argument(0).GetStop(),
// 						"DATE_TRUNC( "+unit+", "+datetime+" )")
// 				}
// 			}
// 		}
// 	}
// }

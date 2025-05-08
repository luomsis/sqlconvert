package ora2pg

import (
	"github/luomsis/sqlconvert/parser"

	"github.com/antlr4-go/antlr/v4"
)

type Ora2pg struct {
	parser.BasePlSqlParserListener
	TokenStreamRewriter *antlr.TokenStreamRewriter
}

func NewOra2pg(tokens antlr.TokenStream) *Ora2pg {
	return &Ora2pg{
		TokenStreamRewriter: antlr.NewTokenStreamRewriter(tokens),
	}
}

func (l *Ora2pg) EnterCreate_table(ctx *parser.Create_tableContext) {
	if ctx.TABLE() != nil {
		l.TokenStreamRewriter.ReplaceDefault(ctx.CREATE().GetSymbol().GetTokenIndex(), ctx.TABLE().GetSymbol().GetTokenIndex(), "CREATE IF NOT EXISTS TABLE")
	}
	l.BasePlSqlParserListener.EnterCreate_table(ctx)
}

func (l *Ora2pg) EnterDatatype(ctx *parser.DatatypeContext) {
	if typeStr := ctx.GetText(); typeStr != "" {
		// 先严格匹配
		switch typeStr {
		case "CLOB", "LONG", "NCLOB":
			l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), "TEXT")
		case "BINARY_FLOAT":
			l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), "REAL")
		case "BINARY_DOUBLE":
			l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), "DOUBLE PRECISION")
		case "INTEGER", "INT", "SMALLINT":
			l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), "DECIMAL(38)")
		case "DATE":
			l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), "TIMESTAMP(0)")
		case "BLOB", "LONG RAW":
			l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), "BYTEA")
		case "BFILE":
			l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), "VARCHAR(255)")
		case "ROWID":
			l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), "CHAR(10)")
		case "SYS_REFCURSOR":
			l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), "REFCURSOR")
		case "XMLTYPE":
			l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), "XML")
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
				l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), replaceStr)

			case "NCHAR":
				replaceStr := "CHAR"
				if ctx.Precision_part() != nil {
					replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + ")"
				}
				l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), replaceStr)

			case "NCHAR VARYING":
				replaceStr := "VARCHAR"
				if ctx.Precision_part() != nil {
					replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + ")"
				}
				l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), replaceStr)
			case "NVARCHAR2", "VARCHAR2":
				replaceStr := "VARCHAR"
				if ctx.Precision_part() != nil {
					replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + ")"
				}
				l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), replaceStr)

			case "FLOAT":
				replaceStr := "DOUBLE PRECISION"
				l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), replaceStr)
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
				l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), replaceStr)
			case "RAW":
				replaceStr := "BYTEA"
				l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), replaceStr)
			case "UROWID":
				replaceStr := "VARCHAR"
				if ctx.Precision_part() != nil {
					replaceStr = replaceStr + "(" + ctx.Precision_part().Numeric(0).GetText() + ")"
				}
				l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), replaceStr)

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
			l.TokenStreamRewriter.ReplaceToken(antlr.DefaultProgramName, ctx.GetStart(), ctx.GetStop(), replaceStr)
		}
	}
	l.BasePlSqlParserListener.EnterDatatype(ctx)
}

func (l *Ora2pg) EnterCreate_function_body(ctx *parser.Create_function_bodyContext) {
	if parameters := ctx.AllParameter(); parameters != nil {
		for i := range parameters {
			if parameters[i].IN(0) != nil && parameters[i].OUT(0) != nil {
				l.TokenStreamRewriter.ReplaceTokenDefault(parameters[i].IN(0).GetSymbol(), parameters[i].OUT(0).GetSymbol(), "INOUT")
			}
		}
	}
	if ctx.RETURN() != nil {
		l.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.RETURN().GetSymbol(), "RETURNS")
	}
	if len(ctx.AllDETERMINISTIC()) != 0 {
		len := len(ctx.AllDETERMINISTIC())
		l.TokenStreamRewriter.DeleteTokenDefault(ctx.AllDETERMINISTIC()[0].GetSymbol(), ctx.AllDETERMINISTIC()[len-1].GetSymbol())
	}
	if ctx.IS() != nil {
		l.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.IS().GetSymbol(), "AS $$")
	}
	if ctx.AS() != nil {
		l.TokenStreamRewriter.ReplaceTokenDefaultPos(ctx.AS().GetSymbol(), "AS $$")
	}
	if ctx.Body() != nil && ctx.Body().Label_name() != nil {
		l.TokenStreamRewriter.DeleteTokenDefault(ctx.Body().Label_name().GetStart(), ctx.Body().Label_name().GetStop())
	}
	l.BasePlSqlParserListener.EnterCreate_function_body(ctx)
}

func (l *Ora2pg) EnterCreate_procedure_body(ctx *parser.Create_procedure_bodyContext) {

}

func (l *Ora2pg) EnterQuery_block(ctx *parser.Query_blockContext) {
	if ctx.From_clause() != nil && ctx.From_clause().Table_ref_list().GetText() == "dual" {
		l.TokenStreamRewriter.DeleteTokenDefault(ctx.From_clause().GetStart(), ctx.From_clause().GetStop())
	}
}
func (l *Ora2pg) EnterGeneral_element_part(ctx *parser.General_element_partContext) {
	if ctx.Id_expression() != nil &&
		ctx.Id_expression().Regular_id() != nil &&
		ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c() != nil &&
		ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().INSTR() != nil {
		if ctx.Function_argument(0) != nil {
			if len(ctx.Function_argument(0).AllArgument()) == 2 {
				a1 := ctx.Function_argument(0).Argument(0)
				a2 := ctx.Function_argument(0).Argument(1)
				l.TokenStreamRewriter.ReplaceTokenDefault(
					ctx.Id_expression().Regular_id().Non_reserved_keywords_pre12c().INSTR().GetSymbol(),
					ctx.Function_argument(0).GetStop(),
					"POSITION("+a2.GetText()+" IN "+a1.GetText()+")")
			} else {
				// unsupport
			}
		}
	}
}

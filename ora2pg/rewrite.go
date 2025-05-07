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
		l.TokenStreamRewriter.ReplaceDefault(ctx.CREATE().GetSymbol().GetTokenIndex(), ctx.TABLE().GetSymbol().GetTokenIndex(), "CREATE MY TABLE")
	}
	l.BasePlSqlParserListener.EnterCreate_table(ctx)
}

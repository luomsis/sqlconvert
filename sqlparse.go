package sqlconvert

import "strings"

var symbols = " _\"'.,;:(){}[]=+-*<>!$~|~`@#%^&/\\\n\r\t"

type Parser struct {
	sourceType    SQLDialect
	scope         *ListWM
	input         []byte
	index         int
	length        int
	line          int
	tokens        []*Token
	PushBackToken *Token
}

func NewParser() *Parser {
	return &Parser{PushBackToken: nil}
}

func (p *Parser) Convert(input []byte) (error, output []byte) {
	// init

	// set app type

	for {
		// get token
		token := p.getNextToken()
		if token == nil {
			break
		}
		p.parse(token)
	}

	// post

	//create output
	return nil, nil
}

func (p *Parser) getNextToken() *Token {
	for p.PushBackToken != nil {
		token := p.PushBackToken
		p.PushBackToken = token.next
		if token.IsBlank() || token.tokenType == TokenComment {
			continue
		}
		return token
	}

	token := newToken()
	for {
		if p.index >= p.length {
			return nil
		}

		// skip space token
		p.skipSpaceTokens()

		// Check for a word first as it can start with special symbol such as _ @
		if p.GetWordToken(token) {
			break
		}

		// Check for a single char token
		if p.GetSingleCharToken(token) {
			break
		}
	}

	return nil
}

func (p *Parser) skipSpaceTokens() {
	for p.index < p.length {
		c := p.input[p.index]
		// check for space character
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			p.index++
			space := newToken()
			space.tokenType = TokenSymbol
			space.content = append(space.content, c)

			p.tokens = append(p.tokens, space)

			if (c == '\n' && p.index < p.length) || (c == '\r' && p.index < p.length && p.input[p.index] != '\n') {
				p.line++
			}
			continue
		}
		break
	}
}

func (p *Parser) ParseComment() bool {
	exists := false
	for {

		p.skipSpaceTokens()

		// signle line comment --
		if p.index < p.length-1 && p.input[p.index] == '-' && p.input[p.index+1] == '-' {
			// use 2 tokens to prepresent the comment
			start := newToken()
			start.tokenType = TokenComment
			start.content = append(start.content, '-', '-')

			p.tokens = append(p.tokens, start)
			p.index += 2
			// mysql requires a blank after --
			// todo:
			len := 0
			for p.index < p.length {
				if p.input[p.index] == '\n' || p.input[p.index] == '\r' {
					break
				}
				len++
				p.index++
			}
			token := newToken()
			token.tokenType = TokenComment
			token.content = p.input[p.index-len : p.index]
			p.tokens = append(p.tokens, token)
			exists = true
			continue

		} else if p.index < p.length-1 && p.input[p.index] == '/' && p.input[p.index+1] == '*' {
			len := 0
			p.index += 2
			for p.index < p.length {
				// go until */
				for p.index < p.length-1 && (p.input[p.index] != '*' || p.input[p.index+1] != '/') {
					len += 2
					p.index += 2
					break
				}

				if p.input[p.index] == '\n' {
					p.line++
				}
				p.index++
			}

			token := newToken()
			token.tokenType = TokenComment
			token.content = p.input[p.index-len-2 : p.index]
			p.tokens = append(p.tokens, token)

			exists = true
			continue
		}
		// informix comment {}
		// Sybase ASA ,Sybase ADS C++ style comment //
		// MySQL single line comment #
		// COBOL single line comment *

		// not a comment
		break
	}

	if exists {
		p.skipSpaceTokens()
	}

	return exists
}

func (p *Parser) GetWordToken(token *Token) bool {
	if token == nil {
		return false
	}
	len := 0
	c := p.input[p.index]
	//todo:
	// Check for a sign for numbers in the first position (sign can go before variable name -num i.e)
	// Skip comment --
	if p.index < p.length-1 && (c == '+' || c == '-') && p.input[p.index+1] != '-' {
		sign := c
		p.index++
		len++

		// Allow spaces follow the sign
		for sign == '-' && p.index < p.length-1 && p.input[p.index] == ' ' {
			p.index++
			len++
		}
	}

	//
	// Identifiers starts as a word but then there is quoted part SCHEMA."TABLE".COL i.e.

	partiallyQuotedIdentifier := false

	for p.index < p.length {
		// check for a comment
		if len == 0 && p.ParseComment() {
			continue
		}

		// check whether we meet a special character allowed in identifiers
		if strings.ContainsRune(symbols, rune(c)) {
			// @variable in SQL Server and MySQL, :new in Oracle trigger, #temp table name in SQL Server
			// * meaning all columns, - in COBOL identifier, label : label name in DB2
			// $ or $$ often used as replacement markers, $ is also allowed in Oracle identifiers
			if c != '_' && c != '.' && c != '@' && c != ':' && c != '#' && c != '*' &&
				c != '-' && c != '"' && c != '[' && c != ' ' && c != '&' && c != '$' {
				break
			}

			// spaces are allowed between identifier parts: table . name
			if c == ' ' {
				identLen := 0
				for i := 0; p.index+i < p.length; i++ {
					if p.input[p.index+i] == ' ' || p.input[p.index+i] == '\t' ||
						p.input[p.index+i] == '\n' || p.input[p.index+i] == '\r' {
						continue
					}

					if p.input[p.index+i] == '.' {
						identLen = i
					}
					break
				}

				// not a multi-part identifier
				if len == 0 || identLen == 0 {
					break
				}
				p.index += identLen
				len += identLen
				continue
			}

			// process sqlserver and sybase

			// * must be after . to not confuse with multiplcation operator
			if p.input[p.index] == '*' && (len == 0 || (len > 0 && p.index > 0 && p.input[p.index-1] != '.')) {
				break
			}

			// check for partially quoted identifier that starts as a word then quoted part follows
			if p.input[p.index] == '"' || p.input[p.index] == '[' {
				if len > 0 && p.index > 0 && p.input[p.index-1] == '.' {
					partiallyQuotedIdentifier = true
				}
				break
			}

			if p.input[p.index] == ':' {
				if (p.index < p.length-1 && p.input[p.index+1] == '=') ||
					p.Source(SQLDB2, SQLTeradata, SQLMySQL) &&
						!p.IsScope(SQLScopeSelectStmt, 0) ||
					(p.index < p.length-1 && p.input[p.index+1] == ':') ||
					(p.index > 0 && p.input[p.index-1] == ':') {
					p.index++
					break
				}
			}

			if p.input[p.index] == '&' && len != 0 {
				break
			}

			right := true
			// Allow - in COBOL only

			// @ must not be followed by a blank or delimiter
			if p.input[p.index] == '@' {
				if p.index == p.length-2 ||
					(p.index < p.length-2 && (p.input[p.index+1] == ' ' || p.input[p.index+1] == '\t' || p.input[p.index+1] == '\r' || p.input[p.index+1] == '\n')) {
					right = false
				}
			}

			if !right {
				break
			}
		}
		p.index++
		len++
	}
	if partiallyQuotedIdentifier {

	}
	return false

}

func (p *Parser) Source(sources ...SQLDialect) bool {
	for _, v := range sources {
		if v == p.sourceType {
			return true
		}
	}
	return false
}

func (p *Parser) IsScope(scope, scope2 SQLClauseScope) bool {
	if p.scope.GetCount() == 0 {
		return false
	}
	for i := p.scope.GetLast(); i != nil; i = i.Next {
		if i.Value1 == scope || i.Value2 == scope2 {
			return true
		}
	}
	return false
}

func (p *Parser) PushBack(token *Token) {
	if token != nil {
		p.PushBackToken = token
	}
}

func (p *Parser) GetSingleCharToken(token *Token) bool {
	return false
}

func (p *Parser) parse(token *Token) {
}

package sqlconvert

import "strings"

var symbols = " _\"'.,;:(){}[]=+-*<>!$~|~`@#%^&/\\\n\r\t"

type Parser struct {
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

	// bool partially_quoted_identifier = false;

	for p.index < p.length {
		// check for a comment
		if len == 0 && p.ParseComment() {
			continue
		}

		// check whether we meet a special character allowed in identifiers
		if strings.ContainsRune(symbols, rune(c)) {
			//
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

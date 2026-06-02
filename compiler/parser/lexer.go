package parser

import (
	"fmt"
	"unicode"
)

// TokenType represents the category of the token.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError
	TokenIdentifier
	TokenKeywordStruct
	TokenKeywordList
	TokenKeywordMap
	TokenKeywordOptional
	TokenLeftBrace
	TokenRightBrace
	TokenColon
	TokenSemicolon
	TokenLeftAngle
	TokenRightAngle
	TokenComma
)

// String returns a human-readable representation of the token type.
func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenError:
		return "Error"
	case TokenIdentifier:
		return "Identifier"
	case TokenKeywordStruct:
		return "struct"
	case TokenKeywordList:
		return "list"
	case TokenKeywordMap:
		return "map"
	case TokenKeywordOptional:
		return "optional"
	case TokenLeftBrace:
		return "{"
	case TokenRightBrace:
		return "}"
	case TokenColon:
		return ":"
	case TokenSemicolon:
		return ";"
	case TokenLeftAngle:
		return "<"
	case TokenRightAngle:
		return ">"
	case TokenComma:
		return ","
	default:
		return "Unknown"
	}
}

// Token represents a lexical token.
type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

// Lexer scans AntiSerial schema files.
type Lexer struct {
	src    []rune
	cursor int
	line   int
	col    int
}

// NewLexer creates a new Lexer instance.
func NewLexer(src string) *Lexer {
	return &Lexer{
		src:  []rune(src),
		line: 1,
		col:  1,
	}
}

// peek returns the character at the current cursor without advancing.
func (l *Lexer) peek() rune {
	if l.cursor >= len(l.src) {
		return 0
	}
	return l.src[l.cursor]
}

// advance returns the current character and moves the cursor forward.
func (l *Lexer) advance() rune {
	if l.cursor >= len(l.src) {
		return 0
	}
	ch := l.src[l.cursor]
	l.cursor++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

// NextToken scans and returns the next Token.
func (l *Lexer) NextToken() Token {
	l.skipWhitespaceAndComments()

	line := l.line
	col := l.col

	ch := l.peek()
	if ch == 0 {
		return Token{Type: TokenEOF, Value: "", Line: line, Col: col}
	}

	// Multi-character tokens (Identifiers and Keywords)
	if unicode.IsLetter(ch) || ch == '_' {
		val := l.readIdentifier()
		var tType TokenType = TokenIdentifier
		switch val {
		case "struct":
			tType = TokenKeywordStruct
		case "list":
			tType = TokenKeywordList
		case "map":
			tType = TokenKeywordMap
		case "optional":
			tType = TokenKeywordOptional
		}
		return Token{Type: tType, Value: val, Line: line, Col: col}
	}

	// Single-character punctuation tokens
	l.advance()
	switch ch {
	case '{':
		return Token{Type: TokenLeftBrace, Value: "{", Line: line, Col: col}
	case '}':
		return Token{Type: TokenRightBrace, Value: "}", Line: line, Col: col}
	case ':':
		return Token{Type: TokenColon, Value: ":", Line: line, Col: col}
	case ';':
		return Token{Type: TokenSemicolon, Value: ";", Line: line, Col: col}
	case '<':
		return Token{Type: TokenLeftAngle, Value: "<", Line: line, Col: col}
	case '>':
		return Token{Type: TokenRightAngle, Value: ">", Line: line, Col: col}
	case ',':
		return Token{Type: TokenComma, Value: ",", Line: line, Col: col}
	default:
		return Token{
			Type:  TokenError,
			Value: fmt.Sprintf("unexpected character: %q", ch),
			Line:  line,
			Col:   col,
		}
	}
}

// skipWhitespaceAndComments skips whitespace characters and single-line comments.
func (l *Lexer) skipWhitespaceAndComments() {
	for {
		ch := l.peek()
		if ch == 0 {
			break
		}
		if unicode.IsSpace(ch) {
			l.advance()
			continue
		}
		// Comment starting with //
		if ch == '/' && l.peekNext() == '/' {
			l.advance() // skip first '/'
			l.advance() // skip second '/'
			for {
				next := l.peek()
				if next == '\n' || next == 0 {
					break
				}
				l.advance()
			}
			continue
		}
		break
	}
}

// peekNext peeks at the character after the current one.
func (l *Lexer) peekNext() rune {
	if l.cursor+1 >= len(l.src) {
		return 0
	}
	return l.src[l.cursor+1]
}

// readIdentifier scans an alphanumeric identifier.
func (l *Lexer) readIdentifier() string {
	start := l.cursor
	// First character is already validated as a letter or underscore
	l.advance()
	for {
		ch := l.peek()
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
			l.advance()
		} else {
			break
		}
	}
	return string(l.src[start:l.cursor])
}

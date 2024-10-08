package main

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode"
)

var (
	trueByte  = []byte("true")
	falseByte = []byte("false")
	nullByte  = []byte("null")
)

// tokenType represents the different YAML tokens
//
//go:generate stringer -type=tokenType
type tokenType int

const (
	INVALID tokenType = iota

	// For indentation and de-indentation
	SPACE tokenType = 1

	// Separator tokens
	COLON  tokenType = 2 // For key-value pairs
	HYPHEN tokenType = 3 // For list elements ("-")

	// Array tokens
	LEFT_SQUARE_BRACKET  = 4
	RIGHT_SQUARE_BRACKET = 5

	// Primitive type tokens
	STRING       tokenType = 6
	FLOAT_NUMBER tokenType = 7
	INT_NUMBER   tokenType = 8
	BOOLEAN      tokenType = 9
	NULL         tokenType = 10

	// Special symbols
	COMMENT tokenType = 11 // Lines beginning with '#'
	NEWLINE tokenType = 12 // Line breaks for separating documents or items

	// Special End of file token
	EOF tokenType = 13
)

// Token containing the value and type of the token, and current pos in the
// input
type Token struct {
	Value string    // Value of the token
	Type  tokenType // The type of the token
	Pos   int       // Position of the token
}

func (t Token) String() string {
	return fmt.Sprintf("Token{Value: %q, Type: %q, Pos: %d}", t.Value, t.Type.String(), t.Pos)
}

// Lexer will read the input and breaks it into tokens
// It will shift from left to right, keeping track of characters
// and move its pos accordingly
type Lexer struct {
	input []byte
	pos   int
}

// nextChar will read the next character from the input, return it
// will return 0 if we have shifted through all the chars in input
// will shift the position to the right of the current char
func (l *Lexer) nextChar() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	ch := l.input[l.pos]
	l.pos++
	return ch
}

// peekChar will read the next character from the input, return it
// will return 0 if we have shifted through all the chars in input
// it will not move the cursor position
func (l *Lexer) peekChar() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	ch := l.input[l.pos]
	return ch
}

func (l *Lexer) NextToken() Token {
	for {
		currentChar := l.nextChar()
		switch currentChar {
		case '\'', '"':
			return l.readQuotedString()
		case ':':
			return Token{Type: COLON, Pos: l.pos - 1}
		case '[':
			return Token{Type: LEFT_SQUARE_BRACKET, Pos: l.pos - 1}
		case ']':
			return Token{Type: RIGHT_SQUARE_BRACKET, Pos: l.pos - 1}
		case '-':
			return Token{Type: HYPHEN, Pos: l.pos - 1}
		case '#':
			return l.readComment()
		case 't', 'f':
			return l.readBoolean()
		case 'n':
			return l.readNull()
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
			return l.readNumber()
		case '\n':
			return Token{Type: NEWLINE, Pos: l.pos - 1}
		case '\t':
			return Token{Type: INVALID, Pos: l.pos - 1, Value: "tabs are not valid within yaml files"}
		case 0:
			return Token{Type: EOF, Pos: l.pos - 1}
		// by default just handle everything as unquoted string
		// except for space chars
		default:
			if unicode.IsSpace(rune(currentChar)) {
				return Token{Type: SPACE, Pos: l.pos}
			}
			return l.readUnquotedString()
		}
	}
}

// readQuotedString will deal with the string with quotes of both types
// i.e. strings with ' and " quotes
func (l *Lexer) readQuotedString() Token {
	start := l.pos - 1 // we need to backtrack to get the first char
	quoteChar := l.input[start]

	// read till the end of string or till we encounter EOF
	for {
		ch := l.nextChar()
		if ch == quoteChar || ch == 0 {
			break
		}
	}

	if l.input[l.pos-1] != quoteChar {
		return Token{Type: INVALID, Pos: start, Value: "unterminated string"}
	}
	return Token{Type: STRING, Value: string(l.input[start:l.pos]), Pos: start}
}

func (l *Lexer) readUnquotedString() Token {
	start := l.pos - 1

	// read till the end of file or till the new line char, or we get comment
	for {
		ch := l.nextChar()
		if ch == 0 || l.isStartOfInlineComment() {
			break
		}
		if l.isStartOfColon() || ch == '\n' {
			l.pos--
			break
		}
	}
	// now remove all the trailing white spaces
	end := l.pos - 1
	if !unicode.IsSpace(rune(l.input[end])) {
		return Token{Type: STRING, Pos: start, Value: string(l.input[start : end+1])}
	}
	for ; end >= 0 && !unicode.IsSpace(rune(l.input[end])); end-- {
	}
	return Token{Type: STRING, Pos: start, Value: string(l.input[start : end+1])}
}

// readNumber will try to read the number from the current position
// will move through the next chars and construct and validate the number
// if found that the number is not valid then it will simply convert it into
// string
func (l *Lexer) readNumber() Token {
	start := l.pos - 1
	numType := INT_NUMBER
	// read till the end of number
	for {
		ch := l.peekChar()

		// change the number type
		if ch == '.' {
			numType = FLOAT_NUMBER
		}

		// check for the end of the line or end of file or end of object, or we get any comment
		if ch == 0 || ch == '\n' || l.isStartOfInlineComment() {
			break
		}
		l.nextChar()
	}

	numStr := string(l.input[start:l.pos])

	// try to parse the number into float, if unsuccessful that means
	// there is some error
	_, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		l.pos = start + 1
		return l.readUnquotedString()
	}
	return Token{Type: numType, Value: numStr, Pos: start}
}

// readBoolean will read through the input bytes and try to parse the booleans
func (l *Lexer) readBoolean() Token {
	start := l.pos - 1
	// read till the end
	for {
		ch := l.peekChar()
		// check for the end of the line or end of file or end of object, or we get comment
		if ch == 0 || ch == '\n' || l.isStartOfInlineComment() {
			break
		}
		l.nextChar()
	}
	boolByte := l.input[start:l.pos]

	if !bytes.Equal(boolByte, trueByte) && !bytes.Equal(boolByte, falseByte) {
		l.pos = start + 1
		return l.readUnquotedString()
	}
	return Token{Type: BOOLEAN, Value: string(boolByte), Pos: start}
}

// readNull will read through the input bytes and try to parse the null
func (l *Lexer) readNull() Token {
	start := l.pos - 1
	// read till the end
	for {
		ch := l.peekChar()
		// check for the end of the line or end of file, or we get comment
		if ch == 0 || ch == '\n' || l.isStartOfInlineComment() || l.isStartOfColon() {
			break
		}
		l.nextChar()
	}
	found := l.input[start:l.pos]
	if !bytes.Equal(found, nullByte) {
		l.pos = start + 1
		return l.readUnquotedString()
	}
	return Token{Type: NULL, Value: string(found), Pos: start}
}

// readComment will just move our lexer to new line, so we can skip the
// commented section
func (l *Lexer) readComment() Token {
	start := l.pos - 1
	// read till the end
	for {
		ch := l.peekChar()
		// check for the end of the line or end of file
		if ch == 0 || ch == '\n' {
			break
		}
		l.nextChar()
	}
	return Token{Type: COMMENT, Pos: start}
}

// isStartOfInlineComment will check for the start of an inline comment
// inline comments are comments at the end of the line instead of starting of a line
// e.g.
//
//	key: value # inline comment
//	# normal comment
func (l *Lexer) isStartOfInlineComment() bool {
	// inline comments will always have a space then '#' character
	// then we can ignore the rest of the chars after the '#' character
	currChar := l.input[l.pos-1]
	if !unicode.IsSpace(rune(currChar)) {
		return false
	}
	return l.peekChar() == '#'
}

func (l Lexer) isStartOfColon() bool {
	// inline comments will always have a space then '#' character
	// then we can ignore the rest of the chars after the '#' character
	currChar := l.input[l.pos-1]
	if currChar != ':' {
		return false
	}
	return unicode.IsSpace(rune(l.peekChar()))
}

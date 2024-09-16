package main

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrDuplicateKey = errors.New("duplicate key")
)

// YAMLObj represents a valid yaml object in the Go world
type YAMLObj struct {
	pairs []KeyValue
	keys  map[string]struct{} // to prevent the duplicate keys
}

// Value implements yamlVal.
func (j YAMLObj) Value() any {
	return j.pairs
}

var _ yamlVal = YAMLObj{} // compile time check

// KeyValue is the key and value of each field of yaml object
type KeyValue struct {
	Key   string
	Value yamlVal
}

func (o *YAMLObj) append(pair KeyValue) error {
	_, ok := o.keys[pair.Key]
	if ok {
		return fmt.Errorf("%w '%s'", ErrDuplicateKey, pair.Key)
	}
	// add the pair key to keys map
	o.keys[pair.Key] = struct{}{}
	o.pairs = append(o.pairs, pair)
	return nil
}

// yamlVal interface must be satisfied by the each primitive value of key value pair of
// yaml object
type yamlVal interface {
	Value() any
}

// yamlString is the representation of string in go
type yamlString string

var _ yamlVal = yamlString("") // compile time check for interface impl

// Value to implement the yamlVal interface
func (s yamlString) Value() any {
	return s
}

// yamlFloat is the representation of floating numbers in go
type yamlFloat float64

var _ yamlVal = yamlFloat(0.0) // compile time check for interface impl

// Value implements yamlVal.
func (j yamlFloat) Value() any {
	return j
}

// yamlInt is the representation of int numbers in go
type yamlInt int

var _ yamlVal = yamlInt(0)

// Value implements yamlVal.
func (j yamlInt) Value() any {
	return j
}

// yamlBool is the representation of boolean in go
type yamlBool bool

var _ yamlVal = yamlBool(false)

// Value implements yamlVal.
func (j yamlBool) Value() any {
	return j
}

// yamlArray representation of array in go
type yamlArray []yamlVal

// Value implements yamlVal.
func (j yamlArray) Value() any {
	return j
}

var _ yamlVal = yamlArray{} // compile time check

// Parser for yaml inputs in byte
type Parser struct {
	lexer              *Lexer
	currToken          Token
	prevToken          *Token
	currentIndentation int
}

func (p Parser) String() string {
	return fmt.Sprintf("Parser{currToken: %s, prevToken: %s}", p.currToken, p.prevToken)
}

// parseError the custom error
type parseError struct {
	Message string
	Pos     int
}

func (e parseError) Error() string {
	return fmt.Sprintf("YAML parse error at position %d: %s", e.Pos, e.Message)
}

func newParseError(msg string, pos int) error {
	return parseError{Message: msg, Pos: pos}
}

var _ error = parseError{}

// NewParser the constructor for the Parser,which initializes the Parser
func NewParser(input []byte) *Parser {
	lex := Lexer{input: input}
	return &Parser{lexer: &lex, currToken: lex.NextToken()}
}

// NextToken the helper function to get the next token from the lexer
// and it sets the currToken to the next token
func (p *Parser) NextToken() {
	if p.prevToken != nil {
		p.currToken = *p.prevToken
		return
	}
	p.currToken = p.lexer.NextToken()
}

func (p *Parser) peekToken() Token {
	p.prevToken = &p.currToken
	return p.lexer.NextToken()
}

// getPos the helper function to get the current token's position
func (p *Parser) getPos() int {
	return p.currToken.Pos
}

func (p *Parser) Parse() (YAMLObj, error) {
	obj := YAMLObj{keys: make(map[string]struct{})}

	if p.currToken.Type == EOF {
		return obj, newParseError("empty files are not valid yaml", p.getPos())
	}

	fmt.Println("before loop")
	fmt.Printf("%s\n", p)
	fmt.Println("in loop")

	for p.currToken.Type != EOF {
		fmt.Printf("%s\n", p)
		if p.currToken.Type != STRING {
			return obj, newParseError("Expected string for key", p.getPos())
		}

		key := p.currToken.Value
		p.NextToken()

		if err := p.expect(COLON, "expected ':' after key declaration"); err != nil {
			return obj, err
		}
		p.NextToken()

		// if err := p.expect(SPACE || NEWLINE, "Expected ' ' after seperator"); err != nil {
		// 	return obj, err
		// }
		if p.currToken.Type != SPACE && p.currToken.Type != NEWLINE {
			return obj, newParseError("Expected ' ' after seperator", p.getPos())
		}
		p.NextToken()
		fmt.Printf("value: %s\n", p)

		val, err := p.parseValue()
		if err != nil {
			return obj, err
		}

		if err := obj.append(KeyValue{Key: key, Value: val}); err != nil {
			return obj, newParseError(err.Error(), p.getPos())
		}

		p.NextToken()
		if p.currToken.Type != NEWLINE && p.currToken.Type != EOF {
			return obj, newParseError("Expected new line after parsing values", p.getPos())
		}
		p.NextToken()
	}

	return obj, nil
}

// Helper function to check the current token type and return an error if it's not expected
func (p Parser) expect(expectedType tokenType, errMsg string) error {
	if p.currToken.Type != expectedType {
		return newParseError(errMsg, p.getPos())
	}
	return nil
}

// parseValue from the current token
func (p *Parser) parseValue() (yamlVal, error) {
	switch p.currToken.Type {
	case STRING:
		value := yamlString(p.currToken.Value)
		return value, nil
	case INT_NUMBER:
		num, err := strconv.Atoi(p.currToken.Value)
		if err != nil {
			return nil, newParseError("expected a number", p.getPos())
		}
		value := yamlInt(num)
		return value, nil
	case FLOAT_NUMBER:
		num, err := strconv.ParseFloat(p.currToken.Value, 64)
		if err != nil {
			return nil, newParseError("Expected a number", p.getPos())
		}
		value := yamlFloat(num)
		return value, nil
	case BOOLEAN:
		bool, err := strconv.ParseBool(p.currToken.Value)
		value := yamlBool(bool)
		if err != nil {
			return value, newParseError("Expected a boolean", p.getPos())
		}
		return value, nil
	case NULL:
		if p.currToken.Value != "null" {
			return nil, newParseError("Expected a null value", p.getPos())
		}
		return nil, nil
	case SPACE:
		return p.handleSpace()
	case NEWLINE:
		return nil, nil
	default:
		return nil, newParseError("Expected value", p.getPos())
	}
}

// Handle SPACE token to track indentation
func (p *Parser) handleSpace() (yamlVal, error) {
	spaceCount := 0
	for p.currToken.Type == SPACE {
		spaceCount++
		p.NextToken()
	}

	// Now we have the number of spaces. We can use this to determine indentation.
	// Based on spaceCount, determine if we're handling a new block (nested list or map).

	if spaceCount > p.currentIndentation {
		// Increase in indentation, start parsing a nested structure (list or map)
		return p.parseIndentedValue()
	} else if spaceCount < p.currentIndentation {
		// Decrease in indentation, we're ending a block. Adjust parser state accordingly.
		p.currentIndentation = spaceCount
		return nil, nil
	}

	// Same level indentation, continue parsing
	return p.parseValue()
}

// Helper to parse indented values (nested structures like lists/maps)
func (p *Parser) parseIndentedValue() (yamlVal, error) {
	p.currentIndentation += 1 // Increase current indentation level

	// Now check what kind of structure it is (list or map) based on the next token
	switch p.currToken.Type {
	case HYPHEN: // Could indicate a YAML list
		return p.parseList()
	case STRING, COLON: // Could indicate a map (key-value pairs)
		return p.Parse()
	default:
		return nil, newParseError("expected list or map after indentation", p.getPos())
	}
}

// Recursively parse a YAML list
func (p *Parser) parseList() (yamlVal, error) {
	var list []yamlVal
	for p.currToken.Type == HYPHEN {
		p.NextToken() // Skip the dash

		item, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		list = append(list, item)

		p.NextToken()
	}
	return yamlArray(list), nil
}

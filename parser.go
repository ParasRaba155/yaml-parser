package main

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrDuplicateKey  = errors.New("duplicate key")
	errEmptyYamlFile = parseError{Message: "empty files are not valid yaml"}
)

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
	// this is the case where we have peeked the next token
	// so swap the peeked token to current token, and set the
	// peeked token to empty
	if p.prevToken != nil {
		p.currToken = *p.prevToken
		p.prevToken = nil
		return
	}
	p.currToken = p.lexer.NextToken()
	fmt.Printf("%s\n", p)
}

func (p *Parser) peekToken() Token {
	p.prevToken = &p.currToken
	fmt.Println("p.peekToken()")
	fmt.Printf("%s\n", p)
	return p.lexer.NextToken()
}

// getPos the helper function to get the current token's position
func (p *Parser) getPos() int {
	return p.currToken.Pos
}

func (p *Parser) Parse() (YAMLObj, error) {
	obj := YAMLObj{keys: make(map[string]struct{})}

	if p.currToken.Type == EOF {
		return obj, errEmptyYamlFile
	}

	for p.currToken.Type != EOF {
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

		val, err := p.parseValue()
		if err != nil {
			return obj, err
		}

		if err := obj.append(KeyValue{Key: key, Value: val}); err != nil {
			return obj, newParseError(err.Error(), p.getPos())
		}

		p.NextToken()
		if p.currToken.Type != NEWLINE && p.currToken.Type != EOF {
			msg := fmt.Sprintf("Expected new line after parsing values, got: %s", p.currToken.Type.String())
			return obj, newParseError(msg, p.getPos())
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
		return p.handleNewLine()
	default:
		return nil, newParseError("Expected value", p.getPos())
	}
}

// Handle SPACE token to track indentation
func (p *Parser) handleSpace() (yamlVal, error) {
	spaceCount := p.getIndentationLevel()

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

func (p *Parser) getIndentationLevel() int {
	spaceCount := 0
	for p.currToken.Type == SPACE {
		spaceCount++
		p.NextToken()
	}
	return spaceCount
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
		return nil, newParseError(fmt.Sprintf("expected list or map after indentation, got: %s", p.currToken.Type.String()), p.getPos())
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

		// get to the next line
		for p.currToken.Type != NEWLINE {
			if p.currToken.Type == EOF {
				break
			}
			p.NextToken()
		}

		if p.currentIndentation > p.getIndentationLevel() {
			break
		}

		if p.currToken.Type == SPACE && p.peekToken().Type != HYPHEN {
			return nil, newParseError("list must have '-' character", p.getPos())
		}
		p.NextToken()
	}
	return yamlArray(list), nil
}

// handleNewLine to track potential map, array or null value
func (p *Parser) handleNewLine() (yamlVal, error) {
	p.NextToken()
	// next token is key and thus current value is NULL
	if p.currToken.Type == STRING {
		return nil, nil
	}
	return p.handleSpace()
}

package main

// tokenType represents the different JSON tokens
//
//go:generate stringer -type=tokenType
type tokenType int

const (
	INVALID tokenType = iota

	// Indentation tokens
	INDENT tokenType = 1
	DEDENT tokenType = 2

	// Separator tokens
	COLON  tokenType = 3 // For key-value pairs
	HYPHEN tokenType = 4 // For list elements ("-")

	// Primitive type tokens
	STRING       tokenType = 5
	FLOAT_NUMBER tokenType = 6
	INT_NUMBER   tokenType = 7
	BOOLEAN      tokenType = 8
	NULL         tokenType = 9

	// Special symbols
	COMMENT tokenType = 10 // Lines beginning with '#'
	NEWLINE tokenType = 11 // Line breaks for separating documents or items

	// Special End of file token
	EOF tokenType = 12
)

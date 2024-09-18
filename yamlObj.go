package main

import "fmt"

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

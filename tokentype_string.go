// Code generated by "stringer -type=tokenType"; DO NOT EDIT.

package main

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[INVALID-0]
	_ = x[INDENT-1]
	_ = x[DEDENT-2]
	_ = x[COLON-3]
	_ = x[HYPHEN-4]
	_ = x[STRING-5]
	_ = x[FLOAT_NUMBER-6]
	_ = x[INT_NUMBER-7]
	_ = x[BOOLEAN-8]
	_ = x[NULL-9]
	_ = x[COMMENT-10]
	_ = x[NEWLINE-11]
	_ = x[EOF-12]
}

const _tokenType_name = "INVALIDINDENTDEDENTCOLONHYPHENSTRINGFLOAT_NUMBERINT_NUMBERBOOLEANNULLCOMMENTNEWLINEEOF"

var _tokenType_index = [...]uint8{0, 7, 13, 19, 24, 30, 36, 48, 58, 65, 69, 76, 83, 86}

func (i tokenType) String() string {
	if i < 0 || i >= tokenType(len(_tokenType_index)-1) {
		return "tokenType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _tokenType_name[_tokenType_index[i]:_tokenType_index[i+1]]
}

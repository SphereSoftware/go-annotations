package registry

import (
	"unicode"
)

const (
	DEFAULT_PARAM = "&"
)

var (
	reserved = map[string]bool{
		"break":       true,
		"default":     true,
		"func":        true,
		"interface":   true,
		"select":      true,
		"case":        true,
		"defer":       true,
		"go":          true,
		"map":         true,
		"struct":      true,
		"chan":        true,
		"else":        true,
		"goto":        true,
		"package":     true,
		"switch":      true,
		"const":       true,
		"fallthrough": true,
		"if":          true,
		"range":       true,
		"type":        true,
		"continue":    true,
		"for":         true,
		"import":      true,
		"return":      true,
		"var":         true,
	}
)

// Returns objects for all found annotations in provided comment
// Objects contain only the map of attribute names and associated values
func FindAnnotations(doc string) []AnnotationDoc {
	var annotations []AnnotationDoc
	chars := []rune(doc)
	n := len(chars)
	for index := 0; index < n; index++ {
		if chars[index] == '@' {
			a, pos := parseAnnotation(chars, index+1, n)
			if a != nil {
				annotations = append(annotations, *a)
				index = pos - 1
			}
		}
	}
	return annotations
}

// Parses one annotation from the position where @ symbol is appeared
// Parameters:
// - array of runes
// - index of symbol '@' in given array
// - length of array
func parseAnnotation(chars []rune, index, n int) (*AnnotationDoc, int) {
	name, pos := parseAnnotationName(chars, index, n)
	if pos == index {
		return nil, pos
	}
	// check for reserved word
	_, found := reserved[name]
	if found {
		panic("Reserved word '" + name + "' can't be used as annotation")
	}
	// parse parameters
	params, pos := parseParameters(chars, pos, n)
	return &AnnotationDoc{name, params}, pos
}

// Parses annotation name starting from provided position.
// Returns found name and index of next char after name.
// Empty name means that no correct name was found
func parseAnnotationName(chars []rune, pos, n int) (string, int) {
	if pos < n {
		start := pos
		if unicode.IsLetter(chars[pos]) {
			pos++
			for pos < n {
				c := chars[pos]
				if unicode.IsLetter(c) ||
					unicode.IsDigit(c) {
					pos++
				} else {
					break
				}
			}
		}
		return string(chars[start:pos]), pos
	} else {
		return "", pos
	}
}

// Parses (optoinal) list of parameters of the annotation.
// Returns the map of parameters names and their attributes
// and the index of the next character after the end of annotation
func parseParameters(chars []rune, pos, n int) (map[string]interface{}, int) {
	params := make(map[string]interface{})
	t, quoted, index := getToken(chars, pos, n)
	if !quoted && t == "(" {
		index = parseOneParamOrList(chars, index, n, params)
		t, quoted, index := getToken(chars, index, n)
		if !quoted && t == ")" {
			return params, index
		} else {
			panic("Annotation parameters are not properly closed by ) " +
				"at '" + string(chars[pos:]) + "'")
		}
	} else {
		return params, pos
	}
}

func parseOneParamOrList(chars []rune, pos, n int, params map[string]interface{}) int {
	t, quoted, index := getToken(chars, pos, n)
	if quoted {
		// it is only one parameter represented as value
		params[DEFAULT_PARAM] = t
		return index
	} else {
		switch t {
		case "@":
			// one element represented as another annotation
			var v *AnnotationDoc
			v, index = parseAnnotation(chars, index, n)
			if v == nil {
				panic("Incorrect parameter format at '" + string(chars[pos:]) + "'")
			}
			params[DEFAULT_PARAM] = *v
			return index
		case "{":
			// one element - array of parameters in {}
			var v interface{}
			v, index = parseArrayParams(chars, index, n)
			params[DEFAULT_PARAM] = v
			return index
		case "}", "(", ")", ",", "=":
			// incorrect parameters list
			panic("Unexpected '" + string(t) + "' in parameters list at '" + string(chars[pos:]) + "'")
		case "":
			panic("Unexpected enf of parameters list at '" + string(chars[pos:]) + "'")
		default:
			// parameter name, '=value[,...]' is expected
			return parseParamList(chars, pos, n, params)
		}
	}
}

func parseArrayParams(chars []rune, pos, n int) (interface{}, int) {
	t, quoted, index := getToken(chars, pos, n)
	if quoted {
		return parseQuotedArrayParams(t, chars, index, n)
	} else {
		switch t {
		case "@":
			// array of tokens
			var v *AnnotationDoc
			v, index = parseAnnotation(chars, index, n)
			if v == nil {
				panic("Incorrect parameter format at '" + string(chars[pos:]) + "'")
			}
			return parseAnnotationArrayParams(v, chars, index, n)
		case "{":
			panic("Array of arrays parameters are not supported")
		case "}", "(", ")", ",", "=":
			panic("Unexpected '" + string(t) + "' in parameters array at '" + string(chars[pos:]) + "'")
		case "":
			panic("Unexpected enf of parameters list at '" + string(chars[pos:]) + "'")
		default:
			// array of numbers
			return parseUnquotedArrayParams(t, chars, index, n)
		}
	}
}

func parseQuotedArrayParams(first string, chars []rune, pos, n int) (interface{}, int) {
	result := []string{first}
	t, quoted, index := getToken(chars, pos, n)
	for !quoted && t == "," {
		t, quoted, index = getToken(chars, index, n)
		if quoted {
			result = append(result, t)
			t, quoted, index = getToken(chars, index, n)
		} else {
			panic("Array of parameters should not contain different types")
		}
	}
	if !quoted && t == "}" {
		return result, index
	}
	panic("No closing '}' in array of parameters at '" + string(chars[pos:]) + "'")
}

func parseAnnotationArrayParams(first *AnnotationDoc, chars []rune, pos, n int) (interface{}, int) {
	result := []AnnotationDoc{*first}
	t, quoted, index := getToken(chars, pos, n)
	for !quoted && t == "," {
		t, quoted, index = getToken(chars, index, n)
		if quoted || t != "@" {
			panic("Array of parameters should not contain different types")
		}
		var v *AnnotationDoc
		v, index = parseAnnotation(chars, index, n)
		if v == nil {
			panic("Incorrect parameter format at '" + string(chars[pos:]) + "'")
		}
		if v.Name == first.Name {
			result = append(result, *v)
			t, quoted, index = getToken(chars, index, n)
		} else {
			panic("Array of parameters should not contain different types")
		}
	}
	if !quoted && t == "}" {
		return result, index
	}
	panic("No closing '}' in array of parameters at '" + string(chars[pos:]) + "'")
}

func parseUnquotedArrayParams(first string, chars []rune, pos, n int) (interface{}, int) {
	result := []string{first}
	t, quoted, index := getToken(chars, pos, n)
	for !quoted && t == "," {
		t, quoted, index = getToken(chars, index, n)
		if quoted {
			panic("Array of parameters should not contain different types")
		} else {
			switch t {
			case "(", ")", ",", "{", "}", "@", "=":
				panic("Array of parameters has incorrect format")
			case "":
				panic("Unexpected enf of parameters list at '" + string(chars[pos:]) + "'")
			}
			result = append(result, t)
			t, quoted, index = getToken(chars, index, n)
		}
	}
	if !quoted && t == "}" {
		return result, index
	}
	panic("No closing '}' in array of parameters at '" + string(chars[pos:]) + "'")
}

func parseParamList(chars []rune, pos, n int, params map[string]interface{}) int {
	moreParams := true
	var name string
	index := pos
	for moreParams {
		name, index = parseParamName(chars, index, n)
		t, quoted, next := getToken(chars, index, n)
		if quoted || t != "=" {
			panic("Equal sign is absent after parameter name at '" + string(chars[pos:]) + "'")
		}
		var v interface{}
		v, next = parseParamValue(chars, next, n)
		params[name] = v
		t, quoted, index = getToken(chars, next, n)
		moreParams = !quoted && t == ","
		if !moreParams {
			index = next
		}
	}
	return index
}

func parseParamName(chars []rune, pos, n int) (string, int) {
	t, quoted, index := getToken(chars, pos, n)
	if quoted {
		panic("Parameter name should not be quoted at '" + string(chars[pos:]) + "'")
	}
	switch t {
	case "{", "}", "(", ")", "@", ",", "=", "":
		panic("parameter name is absent at '" + string(chars[pos:]) + "'")
	}
	return t, index
}

func parseParamValue(chars []rune, pos, n int) (interface{}, int) {
	t, quoted, index := getToken(chars, pos, n)
	if quoted {
		return t, index
	}
	switch t {
	case "@":
		var v *AnnotationDoc
		v, index = parseAnnotation(chars, index, n)
		if v == nil {
			panic("Incorrect parameter value format at '" + string(chars[pos:]) + "'")
		}
		return *v, index
	case "{":
		var v interface{}
		v, index = parseArrayParams(chars, index, n)
		return v, index
	case "}", "(", ")", ",", "=", "":
		panic("Incorrect parameter value at '" + string(chars[pos:]) + "'")
	default:
		return t, index
	}
}

func getToken(chars []rune, start, n int) (string, bool, int) {
	quoted := false
	escaped := false
	var b []rune
	for i := start; i < n; i++ {
		switch chars[i] {
		case '\\':
			if escaped {
				escaped = false
				b = append(b, '\\')
			} else if quoted {
				escaped = true
			} else {
				panic("Unexpected '\\' in annotation attributes in '" + string(chars) + "'")
			}
		case '"':
			if escaped {
				escaped = false
				b = append(b, '"')
			} else if quoted {
				// return unquoted value
				return string(b), true, i + 1
			} else if i == start {
				quoted = true
			} else {
				// return token before quoted value
				return string(b), false, i
			}
		case ' ', '\t', '\r', '\n', '\f':
			if i == start {
				start++
			} else if quoted {
				b = append(b, chars[i])
				escaped = false
			} else {
				return string(b), false, i
			}
		case ',', '=', '(', ')', '{', '}', '@':
			if i == start {
				return string(chars[i]), false, i + 1
			} else if quoted {
				b = append(b, chars[i])
				escaped = false
			} else {
				return string(b), false, i
			}
		default:
			if escaped {
				escaped = false
			}
			b = append(b, chars[i])
		}
	}
	if quoted {
		panic("Unclosed quote in '" + string(chars) + "'")
	}
	return string(b), false, n
}

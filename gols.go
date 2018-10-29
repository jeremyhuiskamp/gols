package gols

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// entry is one layer in a symbol table.
// Don't put lists in here -> runtime panic.
// Are nums allowed?  Can we limit the keys to strings?
type entry map[interface{}]interface{}

// lookup finds the value of a name in an entry.
func (e entry) lookup(name interface{}) (interface{}, bool) {
	if res, ok := e[name]; ok {
		return res, true
	}
	return nil, false
}

// table is a symbol table.
type table []entry

// lookup finds the value of a name in a table.
func (t table) lookup(name interface{}) (interface{}, bool) {
	for _, e := range t {
		if val, ok := e.lookup(name); ok {
			return val, true
		}
	}
	return nil, false
}

func quoteAction(list []interface{}, t table) (interface{}, error) {
	if len(list) != 2 {
		return nil, errors.New("quote must be a list with two elements")
	}
	return list[1], nil
}

func identifierAction(name string, t table) (interface{}, error) {
	if val, ok := t.lookup(name); !ok {
		return nil, fmt.Errorf("unrecognized identifier: %q", name)
	} else {
		return val, nil
	}
}

func lambdaAction(lambda []interface{}, t table) (interface{}, error) {
	if len(lambda) != 3 {
		return nil, errors.New("lambda requires a list with three elements")
	}
	return newLambda(t, lambda[1], lambda[2])
}

func condAction(cond []interface{}, t table) (interface{}, error) {
	lines := cond[1:] // skip "cond" keyword
	for _, line := range lines {
		if cline, ok := line.([]interface{}); !ok {
			return nil, errors.New("cond lines must be lists")
		} else if len(cline) != 2 {
			return nil, errors.New("cond lines must be lists with two elements")
		} else if cline[0] == "else" {
			return meaning(cline[1], t)
		} else {
			matches, err := meaning(cline[0], t)
			if err != nil {
				return nil, err
			}
			// Only place where booleans are significant in
			// the language?
			// Is it an error if the meaning isn't boolean?
			if matches == true {
				return meaning(cline[1], t)
			}
		}
		// do we want to validate the syntax of what comes after
		// a match?  eg, missing else, stuff after an else, etc
	}
	return nil, errors.New("cond must have an else line")
}

func applicationAction(list []interface{}, t table) (interface{}, error) {
	if len(list) == 0 {
		return nil, errors.New("application requires a non-empty list")
	}

	fMeaning, err := meaning(list[0], t)
	if err != nil {
		return nil, err
	}

	type function interface {
		meaning([]interface{}) (interface{}, error)
	}

	f, ok := fMeaning.(function)
	if !ok {
		return nil, fmt.Errorf("unsupported application type: %T", fMeaning)
	}

	args := list[1:]
	argVals := []interface{}{}
	for _, arg := range args {
		argVal, err := meaning(arg, t)
		if err != nil {
			return nil, err
		}
		argVals = append(argVals, argVal)
	}

	return f.meaning(argVals)
}

func meaning(sexp interface{}, t table) (interface{}, error) {
	if list, ok := sexp.([]interface{}); ok {
		if len(list) > 0 {
			if first, ok := list[0].(string); ok {
				switch first {
				case "quote":
					return quoteAction(list, t)
				case "lambda":
					return lambdaAction(list, t)
				case "cond":
					return condAction(list, t)
				}
			}
		}
		// applicationAction is going to have to do quite a
		// lot of error handling!
		return applicationAction(list, t)
	}
	if num, ok := sexp.(uint64); ok {
		return num, nil
	} else if b, ok := sexp.(bool); ok {
		return b, nil
	} else if str, ok := sexp.(string); ok {
		if primitive, ok := nameToPrimitive[str]; ok {
			return primitive, nil
		}
		return identifierAction(str, t)
	}
	return nil, errors.New("unsupported s-expression type")
}

func value(sexp interface{}) (interface{}, error) {
	return meaning(sexp, table([]entry{}))
}

// parsing implementation below copied from http://norvig.com/lispy.html

// tokenize tokenizes an s-expression where only unicode whitespace and
// ()s are considered significant.
func tokenize(src string) []string {
	src = strings.Replace(src, "(", " ( ", -1)
	src = strings.Replace(src, ")", " ) ", -1)
	return strings.Fields(src)
}

// readFromTokens builds an abstract syntax tree from a list of tokens.
// Atoms are either a bool, uint64, or string.  Lists are a []interface{}.
func readFromTokens(tokens []string) (interface{}, []string, error) {
	if len(tokens) == 0 {
		return nil, nil, errors.New("unexpected EOF")
	}

	token := tokens[0]
	tokens = tokens[1:]

	switch token {
	case "(":
		l := []interface{}{} // NB: empty list, not nil
		for len(tokens) > 0 && tokens[0] != ")" {
			sexp, remainder, err := readFromTokens(tokens)
			if err != nil {
				return nil, nil, err
			}
			tokens = remainder
			l = append(l, sexp)
		}
		if len(tokens) < 1 {
			return nil, nil, errors.New("unfinished list")
		}
		return l, tokens[1:], nil
	case ")":
		return nil, nil, errors.New("unexpected )")
	case "#t":
		return true, tokens, nil
	case "#f":
		return false, tokens, nil
	default:
		if num, err := strconv.ParseUint(token, 10, 64); err != nil {
			return token, tokens, nil
		} else {
			return num, tokens, nil
		}
	}
}

// parse tokenizes and builds a syntax tree from an s-expression.
func parse(src string) (interface{}, error) {
	tokens := tokenize(src)
	ast, remainder, err := readFromTokens(tokens)
	if err != nil {
		return nil, err
	}
	if len(remainder) > 0 {
		return nil, errors.New("unexpected trailing tokens")
	}
	return ast, nil
}

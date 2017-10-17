package gols

import (
	"errors"
	"fmt"
	"math"
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

func quoteAction(sexp interface{}, t table) (interface{}, error) {
	if list, ok := sexp.([]interface{}); !ok {
		return nil, errors.New("quote requires a list")
	} else if len(list) != 2 {
		return nil, errors.New("quote must be a list with two elements")
	} else {
		return list[1], nil
	}
}

func identifierAction(sexp interface{}, t table) (interface{}, error) {
	if name, ok := sexp.(string); !ok {
		// is this a bug in the interpreter?
		return nil, errors.New("identifiers must be atoms")
	} else if val, ok := t.lookup(name); !ok {
		return nil, fmt.Errorf("unrecognized identifier: %q", name)
	} else {
		return val, nil
	}
}

func lambdaAction(sexp interface{}, t table) (interface{}, error) {
	lambda, ok := sexp.([]interface{})
	if !ok {
		return nil, errors.New("lambda requires a list")
	} else if len(lambda) != 3 {
		return nil, errors.New("lambda requires a list with three elements")
	}
	// further verification left to the application:
	return []interface{}{
		"non-primitive",
		[]interface{}{
			t,         // hmm, t isn't an s-exp...
			lambda[1], // formals
			lambda[2], // body expression
		},
	}, nil
}

func condAction(sexp interface{}, t table) (interface{}, error) {
	cond, ok := sexp.([]interface{})
	if !ok {
		return nil, errors.New("cond requires a list")
	}
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
			if matches == "#t" {
				return meaning(cline[1], t)
			}
		}
		// do we want to validate the syntax of what comes after
		// a match?  eg, missing else, stuff after an else, etc
	}
	return nil, errors.New("cond must have an else line")
}

func applicationAction(sexp interface{}, t table) (interface{}, error) {
	list, ok := sexp.([]interface{})
	if !ok {
		return nil, errors.New("application requires a list")
	}
	if len(list) == 0 {
		return nil, errors.New("application requires a non-empty list")
	}

	fMeaning, err := meaning(list[0], t)
	if err != nil {
		return nil, err
	}
	// either (primitive foo) or (non-primitive (table formals body))
	f, ok := fMeaning.([]interface{})
	// I think these are bugs in the interpreter?
	if !ok {
		return nil, errors.New("the meaning of a function application must be a list")
	}
	if len(f) != 2 {
		return nil, errors.New(
			"the meaning of a function application must be a " +
				"list with two elements")
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

	if f[0] == "primitive" {
		if name, ok := f[1].(string); !ok {
			return nil, errors.New("name of primitive function must be a string")
		} else {
			return applyPrimitive(name, argVals)
		}
	} else if f[0] == "non-primitive" {
		// f[1] is (table formals body)
		p, ok := f[1].([]interface{})
		if !ok || len(p) != 3 {
			// bug in lambdaAction...
			return nil, errors.New("non-primitive should have three args")
		}
		// how is this different than the table passed to this function?
		t, ok := p[0].(table)
		if !ok {
			return nil, errors.New("non-primitive needs a table")
		}
		formals, ok := p[1].([]interface{})
		if !ok {
			return nil, errors.New("non-primitive requires formals")
		}
		if len(formals) != len(argVals) {
			return nil, errors.New("mismatching number of arguments and parameters")
		}
		e := entry(map[interface{}]interface{}{})
		for i, _ := range formals {
			e[formals[i]] = argVals[i]
		}
		t = append(table([]entry{e}), t...)
		return meaning(p[2], t)
	} else {
		return nil, fmt.Errorf("unsupported application type: %q", f[0])
	}
}

func meaning(sexp interface{}, t table) (interface{}, error) {
	if list, ok := sexp.([]interface{}); ok {
		if len(list) > 0 {
			if first, ok := list[0].(string); ok {
				switch first {
				case "quote":
					return quoteAction(sexp, t)
				case "lambda":
					return lambdaAction(sexp, t)
				case "cond":
					return condAction(sexp, t)
				}
			}
		}
		// applicationAction is going to have to do quite a
		// lot of error handling!
		return applicationAction(sexp, t)
	} else {
		if num, ok := sexp.(uint64); ok {
			return num, nil
		}
		switch sexp {
		case "#t", "#f":
			return sexp, nil
		case "cons", "car", "cdr",
			"null?", "eq?", "atom?",
			"zero?", "add1", "sub1",
			"number?":
			return []interface{}{"primitive", sexp}, nil
		default:
			return identifierAction(sexp, t)
		}
	}
}

func value(sexp interface{}) (interface{}, error) {
	return meaning(sexp, table([]entry{}))
}

// applyPrimitive applies a primitive function.
func applyPrimitive(name string, vals []interface{}) (interface{}, error) {
	bToSexp := func(b bool) interface{} {
		if b {
			return "#t"
		}
		return "#f"
	}

	switch name {
	case "cons":
		if len(vals) != 2 {
			return nil, errors.New("cons takes two arguments")
		} else if to, ok := vals[1].([]interface{}); !ok {
			return nil, errors.New("second argument to cons must be a list")
		} else {
			return append([]interface{}{vals[0]}, to...), nil
		}
	case "car":
		if len(vals) != 1 {
			return nil, errors.New("car takes one argument")
		} else if from, ok := vals[0].([]interface{}); !ok {
			return nil, errors.New("car takes one list")
		} else if len(from) < 1 {
			return nil, errors.New("cannot take car of empty list")
		} else {
			return from[0], nil
		}
	case "cdr":
		if len(vals) != 1 {
			return nil, errors.New("cdr takes one argument")
		} else if from, ok := vals[0].([]interface{}); !ok {
			return nil, errors.New("cdr takes one list")
		} else if len(from) < 1 {
			return nil, errors.New("cannot take cdr of empty list")
		} else {
			return from[1:], nil
		}
	case "null?":
		if len(vals) != 1 {
			return nil, errors.New("null? takes one argument")
		} else if from, ok := vals[0].([]interface{}); !ok {
			return nil, errors.New("null? takes one list")
		} else {
			return bToSexp(len(from) == 0), nil
		}
	case "eq?":
		if len(vals) != 2 {
			return nil, errors.New("eq? takes two arguments")
		} else if first, ok := vals[0].(string); !ok {
			return nil, errors.New("eq? takes two atoms")
		} else if second, ok := vals[1].(string); !ok {
			return nil, errors.New("eq? takes two atoms")
		} else {
			return bToSexp(first == second), nil
		}
	case "atom?":
		if len(vals) != 1 {
			return nil, errors.New("atom? takes one argument")
		}
		// Hmm, support for (primitive x) and (non-privitive x)?
		// The book suggests these are atoms.  How do we hit that case?
		_, ok := vals[0].([]interface{})
		return bToSexp(!ok), nil
	case "zero?":
		if len(vals) != 1 {
			return nil, errors.New("zero? takes one argument")
		} else if num, ok := vals[0].(uint64); !ok {
			return nil, errors.New("zero? takes one number")
		} else {
			return bToSexp(num == 0), nil
		}
	case "add1":
		if len(vals) != 1 {
			return nil, errors.New("add1 takes one argument")
		} else if num, ok := vals[0].(uint64); !ok {
			return nil, errors.New("add1 takes one number")
		} else if num == math.MaxUint64 {
			return nil, errors.New("add1 would cause overflow")
		} else {
			return num + 1, nil
		}
	case "sub1":
		if len(vals) != 1 {
			return nil, errors.New("sub1 takes one argument")
		} else if num, ok := vals[0].(uint64); !ok {
			return nil, errors.New("sub1 takes one number")
		} else if num == 0 {
			return nil, errors.New("sub1 would cause underflow")
		} else {
			return num - 1, nil
		}
	case "number?":
		if len(vals) != 1 {
			return nil, errors.New("number? takes one argument")
		}
		_, ok := vals[0].(uint64)
		return bToSexp(ok), nil
	default:
		return nil, fmt.Errorf("unknown primitive: %q", name)
	}
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
// Atoms are either a string or a uint64.  Lists are a []interface{}.
// TODO: consider #f and #t as bool types?
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

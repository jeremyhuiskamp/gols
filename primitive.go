package gols

import (
	"errors"
	"math"
)

type functionFunc func([]interface{}) (interface{}, error)

func (f functionFunc) meaning(args []interface{}) (interface{}, error) {
	return f(args)
}

// This looks a lot like a symbol table entry, and should probably actually be
// the first layer.  The book doesn't implement it this way though, and
// prevents code from binding new expressions to these names.  This seems to be
// different than what other scheme implementations do.
var nameToPrimitive = map[string]functionFunc{
	"cons":    cons,
	"car":     car,
	"cdr":     cdr,
	"null?":   isNull,
	"eq?":     isEq,
	"atom?":   isAtom,
	"zero?":   isZero,
	"add1":    add1,
	"sub1":    sub1,
	"number?": isNumber,
}

func bToSexp(b bool) interface{} {
	if b {
		return "#t"
	}
	return "#f"
}

func cons(vals []interface{}) (interface{}, error) {
	if len(vals) != 2 {
		return nil, errors.New("cons takes two arguments")
	} else if to, ok := vals[1].([]interface{}); !ok {
		return nil, errors.New("second argument to cons must be a list")
	} else {
		return append([]interface{}{vals[0]}, to...), nil
	}
}

func car(vals []interface{}) (interface{}, error) {
	if len(vals) != 1 {
		return nil, errors.New("car takes one argument")
	} else if from, ok := vals[0].([]interface{}); !ok {
		return nil, errors.New("car takes one list")
	} else if len(from) < 1 {
		return nil, errors.New("cannot take car of empty list")
	} else {
		return from[0], nil
	}
}

func cdr(vals []interface{}) (interface{}, error) {
	if len(vals) != 1 {
		return nil, errors.New("cdr takes one argument")
	} else if from, ok := vals[0].([]interface{}); !ok {
		return nil, errors.New("cdr takes one list")
	} else if len(from) < 1 {
		return nil, errors.New("cannot take cdr of empty list")
	} else {
		return from[1:], nil
	}
}

func isNull(vals []interface{}) (interface{}, error) {
	if len(vals) != 1 {
		return nil, errors.New("null? takes one argument")
	} else if from, ok := vals[0].([]interface{}); !ok {
		return nil, errors.New("null? takes one list")
	} else {
		return bToSexp(len(from) == 0), nil
	}
}

func isEq(vals []interface{}) (interface{}, error) {
	if len(vals) != 2 {
		return nil, errors.New("eq? takes two arguments")
	} else if first, ok := vals[0].(string); !ok {
		return nil, errors.New("eq? takes two atoms")
	} else if second, ok := vals[1].(string); !ok {
		return nil, errors.New("eq? takes two atoms")
	} else {
		return bToSexp(first == second), nil
	}
}

func isAtom(vals []interface{}) (interface{}, error) {
	if len(vals) != 1 {
		return nil, errors.New("atom? takes one argument")
	}
	// Hmm, support for (primitive x) and (non-privitive x)?
	// The book suggests these are atoms.  How do we hit that case?
	_, ok := vals[0].([]interface{})
	return bToSexp(!ok), nil
}

func isZero(vals []interface{}) (interface{}, error) {
	if len(vals) != 1 {
		return nil, errors.New("zero? takes one argument")
	} else if num, ok := vals[0].(uint64); !ok {
		return nil, errors.New("zero? takes one number")
	} else {
		return bToSexp(num == 0), nil
	}
}

func add1(vals []interface{}) (interface{}, error) {
	if len(vals) != 1 {
		return nil, errors.New("add1 takes one argument")
	} else if num, ok := vals[0].(uint64); !ok {
		return nil, errors.New("add1 takes one number")
	} else if num == math.MaxUint64 {
		return nil, errors.New("add1 would cause overflow")
	} else {
		return num + 1, nil
	}
}

func sub1(vals []interface{}) (interface{}, error) {
	if len(vals) != 1 {
		return nil, errors.New("sub1 takes one argument")
	} else if num, ok := vals[0].(uint64); !ok {
		return nil, errors.New("sub1 takes one number")
	} else if num == 0 {
		return nil, errors.New("sub1 would cause underflow")
	} else {
		return num - 1, nil
	}
}

func isNumber(vals []interface{}) (interface{}, error) {
	if len(vals) != 1 {
		return nil, errors.New("number? takes one argument")
	}
	_, ok := vals[0].(uint64)
	return bToSexp(ok), nil
}

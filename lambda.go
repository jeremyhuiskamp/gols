package gols

import "errors"

// lambda is the partially parsed representation of a lambda expression.  It
// can be evaluated against a set of arguments.
type lambda struct {
	// t is the symbol table at the point where the lambda is defined.
	t       table
	formals []string
	body    interface{}
}

// newLambda creates a new lambda from a surrounding symbol table, a formals
// s-expression and a body s-expression.
func newLambda(t table, formals interface{}, body interface{}) (*lambda, error) {
	formalsL, ok := formals.([]interface{})
	if !ok {
		return nil, errors.New("lambda formals must be a list")
	}
	var formalsStr []string
	for _, formal := range formalsL {
		formalStr, ok := formal.(string)
		if !ok {
			return nil, errors.New("lambda formals must be symbols")
		}
		formalsStr = append(formalsStr, formalStr)
	}

	return &lambda{t, formalsStr, body}, nil
}

// meaning evaluates the meaning of the lambda against the given arguments
func (l *lambda) meaning(args []interface{}) (interface{}, error) {
	if len(args) != len(l.formals) {
		return nil, errors.New("wrong number of arguments to lambda")
	}
	e := entry(map[interface{}]interface{}{})
	for i := range l.formals {
		e[l.formals[i]] = args[i]
	}
	return meaning(l.body, append(table([]entry{e}), l.t...))
}

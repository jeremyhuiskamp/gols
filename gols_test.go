package gols

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMeaningValid(t *testing.T) {
	for in, out := range map[string]string{
		"1":             "1",
		"(quote (foo))": "(foo)",
		"(quote foo)":   "foo",

		"(atom? 1)":          "#t",
		"(atom? (quote x))":  "#t",
		"(atom? (quote ()))": "#f",

		"(null? (quote ()))":  "#t",
		"(null? (quote (x)))": "#f",

		"(eq? (quote x) (quote x))": "#t",
		"(eq? (quote x) (quote y))": "#f",

		"(number? 1)":          "#t",
		"(number? (quote x))":  "#f",
		"(number? (quote ()))": "#f",

		"(zero? 0)": "#t",
		"(zero? 1)": "#f",

		"(add1 0)": "1",
		"(sub1 1)": "0",

		"(cons 1 (quote ()))":            "(1)",
		"(cons 1 (quote (2)))":           "(1 2)",
		"(cons (quote (1)) (quote (2)))": "((1) 2)",

		"(car (quote (1 2)))":         "1",
		"(car (quote ((1) 2)))":       "(1)",
		"(car (car (quote ((1) 2))))": "1",

		"(cdr (quote (1)))":           "()",
		"(cdr (quote (1 2)))":         "(2)",
		"(cdr (quote (1 2 3)))":       "(2 3)",
		"(cdr (cdr (quote (1 2 3))))": "(3)",

		"(cond (#t 1) (else 2))":             "1",
		"(cond (#f 1) (else 2))":             "2",
		"(cond ((number? 1) #t) (else #f))":  "#t",
		"(cond ((number? #t) #t) (else #f))": "#f",

		"((lambda (f a) (f a)) (lambda (a) (add1 a)) 2)": "3",

		`(((lambda (le)
	 	    ((lambda (f) (f f))
		       (lambda (f) (le (lambda (x) ((f f) x))))))
		   (lambda (length)
		     (lambda (l)
		       (cond
		         ((null? l) 0)
		         (else (add1 (length (cdr l))))))))
		  (quote (a b c d)))`: "4",
	} {
		t.Log(in)
		inExp, err := parse(in)
		if err != nil {
			t.Fatalf("parse error: %s", err) // test bug
		}

		outExp, err := value(inExp)
		if err != nil {
			t.Fatalf("evaluation error: %s", err) // test bug
		}

		got := sexpToString(outExp)
		if got != out {
			t.Fatalf("unexpected output: %q", got)
		}
	}
}

func TestMeaningInvalid(t *testing.T) {
	// This is an attempt to hit all hittable error handling.
	// It's verified mostly by code coverage because it's not
	// clear exactly how to assert on what error has been
	// returned.
	// Not all error handling is hittable, because there's some
	// safety checks that would represent bugs in the interpreter.
	for _, in := range []string{
		"x", // unknown identifier

		"(quote)",     // nothing to quote
		"(quote a b)", // too many things to quote

		"(lambda)",         // no formals
		"(lambda (x))",     // no body
		"(lambda (x) 2 3)", // too much body

		"(cond 1)",       // line not a list
		"(cond (1))",     // not enough elements in line
		"(cond (1 2 3))", // too many elements in line
		"(cond ((x) 2))", // meaning of condition undefined
		"(cond (#f 2))",  // no match

		"()",                   // application of nothing
		"(1)",                  // application of a non-function
		"((quote 1))",          // application of a non-function
		"(null? x)",            // undefined arguments to function
		"((lambda 1 2) 3)",     // formals not a list
		"((lambda (x) 2) 3 4)", // too many arguments

		"(cons 1)",   // need a second argument
		"(cons 1 2)", // second argument must be a list

		"(car)",               // no arguments
		"(car (quote (1)) 2)", // too many arguments
		"(car 1)",             // argument not a list
		"(car (quote ()))",    // empty list

		"(cdr)",               // no arguments
		"(cdr (quote (1)) 2)", // too many arguments
		"(cdr 1)",             // argument not a list
		"(cdr (quote ()))",    // empty list

		"(null?)",     // no arguments
		"(null? 1 2)", // too many arguments
		"(null? 1)",   // argument not a list

		"(eq?)",                       // no arguments
		"(eq? 1)",                     // not enough arguments
		"(eq? 1 2 3)",                 // too many arguments
		"(eq? 1 2)",                   // arguments are numeric
		"(eq? #f 2)",                  // arguments are numeric
		"(eq? (quote ()) (quote ()))", // arguments are not atoms

		"(atom?)",     // no arugments
		"(atom? 1 2)", // too many arguments

		"(zero?)",     // no arguments
		"(zero? 1 2)", // too many arguments
		"(zero? #f)",  // argument not a number

		"(add1)",                      // no arguments
		"(add1 2 3)",                  // too many arguments
		"(add1 (quote ()))",           // argument not a number
		"(add1 18446744073709551615)", // overflow

		"(sub1)",     // no arguments
		"(sub1 2 3)", // too many arguments
		"(sub1 #f)",  // argument not a number
		"(sub1 0)",   // underflow

		"(number?)",     // no arguments
		"(number? 2 3)", // too many arguments

		"(wat 2)",         // unknown function name
		"(#f 2)",          // non-executable function
		"(1 2)",           // non-executable function
		"((quote foo) 3)", // non-executable function

		"((lambda (1) (2)) 3)",   // non-symbol formal
		"((lambda (x) (1)) 2 3)", // wrong number of arguments
	} {
		inExp, err := parse(in)
		if err != nil {
			t.Fatalf("parse error: %s", err) // test bug
		}

		_, err = value(inExp)
		if err == nil {
			t.Fatalf("got no error for input %s", in)
		} else {
			t.Logf("%s -> %s", in, err)
		}
	}
}

func TestTokenize(t *testing.T) {
	for _, test := range []struct {
		src string
		exp []string
	}{
		{
			"",
			[]string{},
		},
		{
			"x",
			[]string{"x"},
		},
		{
			" x\t 1 \n x",
			[]string{"x", "1", "x"},
		},
		{
			"()",
			[]string{"(", ")"},
		},
		{
			"(x)",
			[]string{"(", "x", ")"},
		},
	} {
		name := fmt.Sprintf("test: %q", test.src)
		t.Run(name, func(t *testing.T) {
			got := tokenize(test.src)
			if !reflect.DeepEqual(got, test.exp) {
				t.Fatalf("unexpected parsed tokens: %#v", got)
			}
		})
	}
}

func TestReadFromTokens(t *testing.T) {
	for _, test := range []struct {
		tokens    []string
		err       bool
		tree      interface{}
		remainder []string
	}{
		{
			tokens: []string{},
			err:    true,
		},
		{
			tokens: []string{"x"},
			tree:   "x",
		},
		{
			tokens: []string{"1"},
			tree:   uint64(1),
		},
		{
			tokens: []string{"(", "x", ")"},
			tree:   []interface{}{"x"},
		},
		{
			tokens: []string{"(", ")"},
			tree:   []interface{}{},
		},
		{
			tokens: []string{"(", "(", ")", "x", ")"},
			tree:   []interface{}{[]interface{}{}, "x"},
		},
		{
			tokens: []string{"(", "x", "1", "*", ")"},
			tree:   []interface{}{"x", uint64(1), "*"},
		},
		{
			tokens: []string{"(", "x"},
			err:    true,
		},
		{
			tokens: []string{")", "x"},
			err:    true,
		},
		{
			tokens: []string{"(", "(", "x"},
			err:    true,
		},
		{
			tokens:    []string{"(", "x", ")", "x"},
			tree:      []interface{}{"x"},
			remainder: []string{"x"},
		},
		{
			tokens:    []string{"x", "x"},
			tree:      "x",
			remainder: []string{"x"},
		},
	} {
		name := fmt.Sprintf("test: %q", test.tokens)
		t.Run(name, func(t *testing.T) {
			if test.remainder == nil {
				test.remainder = []string{}
			}
			got, remainder, err := readFromTokens(test.tokens)
			if test.err {
				if err == nil {
					t.Fatal("didn't get expected error")
				}
			} else if err != nil {
				t.Fatal(err)
			} else if !reflect.DeepEqual(got, test.tree) {
				t.Fatalf("unexpected tree: %#v", got)
			} else if !reflect.DeepEqual(remainder, test.remainder) {
				t.Fatalf("unexpected remainder: %#v", remainder)
			}
		})
	}
}

func TestParse(t *testing.T) {
	for _, test := range []struct {
		src  string
		err  bool
		tree interface{}
	}{
		{
			src:  "x",
			tree: "x",
		},
		{
			src: "(x",
			err: true, // can't readFromTokens
		},
		{
			src: "x x",
			err: true, // trailing tokens
		},
		{
			src: "(begin (define r 10) (* pi (* r r)))",
			tree: []interface{}{
				"begin",
				[]interface{}{
					"define",
					"r",
					uint64(10),
				},
				[]interface{}{
					"*",
					"pi",
					[]interface{}{
						"*",
						"r",
						"r",
					},
				},
			},
		},
	} {
		name := fmt.Sprintf("test: %q", test.src)
		t.Run(name, func(t *testing.T) {
			got, err := parse(test.src)
			if test.err {
				if err == nil {
					t.Fatal("didn't get expected error")
				}
			} else if err != nil {
				t.Fatal(err)
			} else if !reflect.DeepEqual(got, test.tree) {
				t.Fatalf("unexpected tree: %#v", got)
			}
		})
	}
}

func sexpToString(sexp interface{}) string {
	res := ""
	if l, ok := sexp.([]interface{}); ok {
		res += "("
		for i, item := range l {
			res += sexpToString(item)
			if i < len(l)-1 {
				res += " "
			}
		}
		res += ")"
	} else if i, ok := sexp.(uint64); ok {
		res += fmt.Sprintf("%d", i)
	} else {
		res += fmt.Sprintf("%s", sexp)
	}
	return res
}

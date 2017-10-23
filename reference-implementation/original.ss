; This is a literal copy from Chapter 10 of the Little Schemer, intended for
; use in compatibility tests.

(define atom?
  (lambda (x)
    (and (not (pair? x)) (not (null? x)))))

(define first
  (lambda (p)
    (car p)))

(define second
  (lambda (p)
    (car (cdr p))))

(define third
  (lambda (p)
    (car (cdr (cdr p)))))

(define build
  (lambda (s1 s2)
    (cons s1 (cons s2 (quote ())))))

(define new-entry build)

(define extend-table cons)

(define lookup-in-entry-help
  (lambda (name names values entry-f)
    (cond
      ((null? names) (entry-f name))
      ((eq? (car names) name)
       (car values))
      (else 
	(lookup-in-entry-help 
	  name
	  (cdr names)
	  (cdr values)
	  entry-f)))))

(define lookup-in-entry
  (lambda (name entry entry-f)
    (lookup-in-entry-help name
			  (first entry)
			  (second entry)
			  entry-f)))

(define lookup-in-table
  (lambda (name table table-f)
    (cond
      ((null? table) (table-f name))
      (else (lookup-in-entry name
			     (car table)
			     (lambda (name)
			       (lookup-in-table name
						(cdr table)
						table-f)))))))

(define initial-table
  (lambda (name)
    (car (quote ()))))

(define *identifier
  (lambda (e table)
    (lookup-in-table e table initial-table)))

(define text-of second)

(define *quote
  (lambda (e table)
    (text-of e)))

(define *const
  (lambda (e table)
    (cond
      ((number? e) e)
      ((eq? e #t) #t)
      ((eq? e #f) #f)
      (else (build (quote primitive) e)))))

(define *lambda
  (lambda (e table)
    (build (quote non-primitive)
	   (cons table (cdr e)))))

(define table-of first)

(define formals-of second)

(define body-of third)

(define else?
  (lambda (x)
    (cond
      ((atom? x) (eq? x (quote else)))
      (else #f))))

(define question-of first)

(define answer-of second)

(define evcon
  (lambda (lines table)
    (cond
      ((else? (question-of (car lines)))
       (meaning (answer-of (car lines))
		table))
      ((meaning (question-of (car lines))
		table)
       (meaning (answer-of (car lines))
		table))
      (else (evcon (cdr lines) table)))))

(define cond-lines-of cdr)

(define *cond
  (lambda (e table)
    (evcon (cond-lines-of e) table)))

(define evlis
  (lambda (args table)
    (cond
      ((null? args) (quote ()))
      (else
	(cons (meaning (car args) table)
	      (evlis (cdr args) table))))))

(define function-of car)

(define arguments-of cdr)

(define primitive?
  (lambda (l)
    (eq? (first l) (quote primitive))))

(define non-primitive?
  (lambda (l)
    (eq? (first l) (quote non-primitive))))

(define :atom?
  (lambda (x)
    (cond
      ((atom? x) #t)
      ((null? x) #f)
      ((eq? (car x) (quote primitive)) #t)
      ((eq? (car x) (quote non-primitive)) #f)
      (else #f))))

(define apply-primitive
  (lambda (name vals)
    (cond
      ((eq? name (quote cons))
       (cons (first vals) (second vals)))
      ((eq? name (quote car))
       (car (first vals)))
      ((eq? name (quote cdr))
       (cdr (first vals)))
      ((eq? name (quote null?))
       (null? (first vals)))
      ((eq? name (quote eq?))
       (eq? (first vals) (second vals)))
      ((eq? name (quote atom?))
       (:atom? (first vals)))
      ((eq? name (quote zero?))
       (zero? (first vals)))
      ((eq? name (quote add1))
       (+ 1 (first vals)))
      ((eq? name (quote sub1))
       (- 1 (first vals)))
      ((eq? name (quote number?))
       (number? (first vals))))))

(define apply-closure
  (lambda (closure vals)
    (meaning (body-of closure)
	     (extend-table
	       (new-entry
		 (formals-of closure)
		 vals)
	       (table-of closure)))))

(define apply
  (lambda (fun vals)
    (cond
      ((primitive? fun)
       (apply-primitive
	 (second fun) vals))
      ((non-primitive? fun)
       (apply-closure
	 (second fun) vals)))))

(define *application
  (lambda (e table)
    (apply
      (meaning (function-of e) table)
      (evlis (arguments-of e) table))))

(define atom-to-action
  (lambda (e)
    (cond
      ((number? e) *const)
      ((eq? e #t) *const)
      ((eq? e #f) *const)
      ((eq? e (quote cons)) *const)
      ((eq? e (quote car)) *const)
      ((eq? e (quote cdr)) *const)
      ((eq? e (quote null?)) *const)
      ((eq? e (quote eq?)) *const)
      ((eq? e (quote atom?)) *const)
      ((eq? e (quote zero?)) *const)
      ((eq? e (quote add1)) *const)
      ((eq? e (quote sub1)) *const)
      ((eq? e (quote number?)) *const)
      (else *identifier))))

(define list-to-action
  (lambda (e)
    (cond
      ((atom? (car e))
       (cond
	 ((eq? (car e) (quote quote))
	  *quote)
	 ((eq? (car e) (quote lambda))
	  *lambda)
	 ((eq? (car e) (quote cond))
	  *cond)
	 (else *application)))
       (else *application))))

(define expression-to-action
  (lambda (e)
    (cond
      ((atom? e) (atom-to-action e))
      (else (list-to-action e)))))

(define meaning
  (lambda (e table)
    ((expression-to-action e) e table)))

(define value
  (lambda (e)
    (meaning e (quote ()))))

; A mini test-suite:

; evaluate-test returns (result expected expression) where result is whether
; the value of the expression matches the expectation.
(define evaluate-test
  (lambda (expected expression)
    (cons (eq? expected (value expression))
	  (cons expected
		(cons expression
		      (quote ()))))))

; evaluate-tests takes a list of test-cases in the form (expected expression)
; and maps them to evaluate-test
(define evaluate-tests
  (lambda (test-cases)
    (cond
      ((null? test-cases) (quote ()))
      (else
	(cons 
	  (evaluate-test (first (car test-cases)) (second (car test-cases)))
	  (evaluate-tests (cdr test-cases)))))))

; print-results just prints each item in the list on a separate line
(define print-results
  (lambda (results)
    (cond
      ((null? results) (quote ()))
      (else
	(display (car results))
	(newline)
	(print-results (cdr results))))))

; #f in the left-most column in the output means that a test failed.
(print-results
  (evaluate-tests
    (quote (
       (#t (number? 1))
       (#f (number? (quote x)))
       (1 (add1 0))
       (3 ((lambda (f a) (f a)) add1 2))
       (3 ((lambda (f a) (f a)) (lambda (a) (add1 a)) 2))
       (4 (((lambda (le)
	      ((lambda (f) (f f))
	       (lambda (f) (le (lambda (x) ((f f) x))))))
	    (lambda (length)
	      (lambda (l)
		(cond
		  ((null? l) 0)
		  (else (add1 (length (cdr l))))))))
	   (quote (a b c d))))
       (a ((lambda (car) (car (quote (a b c)))) cdr))
       (1 ((lambda (x) x) 1))
       (#t (eq? #t #t)) ; eq? against booleans
       (#t (eq? #f #f))
       (#f (eq? #f #t))
       (#t (eq? (quote x) (quote x)))
       ))))

// Copyright 2025 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package jsonpath

import "fmt"

// type JsonpathExpr = tree.Expr
type JsonpathExpr interface {
	fmt.Stringer
}

// Identical to JsonpathExpr for now.
type Accessor interface {
	JsonpathExpr
}

// Satisfies JsonpathExpr interface.
type Jsonpath struct {
	Query  Query
	Strict bool
}

func (j Jsonpath) String() string {
	var mode string
	if j.Strict {
		mode = "strict "
	}
	return mode + j.Query.String()
}

// Satisfies JsonpathExpr interface.
type Query struct {
	Accessors []Accessor
}

func (q Query) String() string {
	var s string
	for _, accessor := range q.Accessors {
		s += accessor.String()
	}
	return s
}

// Satisfies Accessor interface.
type Root struct{}

func (r Root) String() string { return "$" }

// Satisfies Accessor interface.
type Key struct {
	Key string
}

func (k Key) String() string { return "." + k.Key }

// Satisfies Accessor interface.
type Wildcard struct{}

func (w Wildcard) String() string { return "[*]" }

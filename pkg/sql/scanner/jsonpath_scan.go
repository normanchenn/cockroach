// Copyright 2025 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package scanner

import (
	sqllexbase "github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/util/jsonpath/parser/lexbase"
)

// JSONPathScanner is a scanner with a jsonpath-specific scan function
type JSONPathScanner struct {
	Scanner
}

func (s *JSONPathScanner) Scan(lval ScanSymType) {
	ch, skipWhiteSpace := s.scanSetup(lval)
	if skipWhiteSpace {
		return
	}

	if isIdentStart(ch) {
		s.scanIdent(lval)
		return
	}
}

func isIdentStart(ch int) bool {
	return (ch >= 'A' && ch <= 'Z') ||
		(ch >= 'a' && ch <= 'z') ||
		sqllexbase.IsDigit(ch)
}

func (s *JSONPathScanner) scanIdent(lval ScanSymType) {
	s.lowerCaseAndNormalizeIdent(lval, isIdentStart)
	lval.SetID(lexbase.GetKeywordID(lval.Str()))
}

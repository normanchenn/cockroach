// Copyright 2025 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package scanner

import (
	"fmt"
	"go/constant"
	"go/token"

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

	// check if it can be turned into a number; if not, then try to scan an ident.
	// now 1a will be scanned as a IDENT instead of an error.
	if sqllexbase.IsDigit(ch) && s.scanNumber(lval, ch) {
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

func (s *JSONPathScanner) scanNumber(lval ScanSymType, ch int) bool {
	start := s.pos - 1
	hasDecimal := ch == '.'
	hasExponent := false

	for {
		ch := s.peek()
		if sqllexbase.IsDigit(ch) {
			s.pos++
			continue
		}
		// skip hex check
		if ch == '.' {
			if hasDecimal || hasExponent {
				break
			}
			s.pos++
			if s.peek() == '.' {
				// Found ".." while scanning a number: back up to the end of the
				// integer.
				s.pos--
				break
			}
			hasDecimal = true
			continue
		}
		if ch == 'e' || ch == 'E' {
			if hasExponent {
				break
			}
			hasExponent = true
			s.pos++
			ch = s.peek()
			if ch == '-' || ch == '+' {
				s.pos++
			}
			ch = s.peek()
			if !sqllexbase.IsDigit(ch) {
				lval.SetID(lexbase.ERROR)
				lval.SetStr("invalid floating point literal")
				return true
			}
			continue
		}
		break
	}

	// Disallow identifier after numerical constants e.g. "124foo".
	if isIdentStart(s.peek()) {
		lval.SetID(lexbase.ERROR)
		lval.SetStr(fmt.Sprintf("trailing junk after numeric literal at or near %q", s.in[start:s.pos+1]))
		return false
	}

	lval.SetStr(s.in[start:s.pos])
	if hasDecimal || hasExponent {
		lval.SetID(lexbase.FCONST)
		floatConst := constant.MakeFromLiteral(lval.Str(), token.FLOAT, 0)
		if floatConst.Kind() == constant.Unknown {
			lval.SetID(lexbase.ERROR)
			lval.SetStr(fmt.Sprintf("could not make constant float from literal %q", lval.Str()))
			return true
		}
		// lval.SetUnionVal(NewNumValFn(floatConst, lval.Str(), false /* negative */))
	} else {
		// Strip off leading zeros from decimal literals so that
		// constant.MakeFromLiteral doesn't inappropriately interpret the
		// string as an octal literal. Note: we can't use strings.TrimLeft
		// here, because it will truncate '0' to ''.
		for len(lval.Str()) > 1 && lval.Str()[0] == '0' {
			lval.SetStr(lval.Str()[1:])
		}

		// lval.SetID(lexbase.ICONST)
		lval.SetID(lexbase.INTEGER)
		intConst := constant.MakeFromLiteral(lval.Str(), token.INT, 0)
		if intConst.Kind() == constant.Unknown {
			lval.SetID(lexbase.ERROR)
			lval.SetStr(fmt.Sprintf("could not make constant int from literal %q", lval.Str()))
			return true
		}
		// lval.SetUnionVal(NewNumValFn(intConst, lval.Str(), false /* negative */))
	}
	return true
}

// Copyright 2025 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package parser

import (
	"github.com/cockroachdb/cockroach/pkg/sql/scanner"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/jsonpath"
	"github.com/cockroachdb/errors"
)

var defaultNakedIntType = types.Int

func init() {

}

type Parser struct {
	scanner    scanner.JSONPathScanner
	lexer      lexer
	parserImpl jsonpathParserImpl
}

func (p *Parser) scan() (query string, tokens []jsonpathSymType, done bool) {
	var lval jsonpathSymType

	p.scanner.Scan(&lval)
	if lval.id == 0 {
		return "", nil, true
	}

	startPos := lval.pos

	lval.pos = 0
	tokens = append(tokens, lval)
	var posBeforeScan int
	for {
		if lval.id == ERROR {
			return p.scanner.In()[startPos:], tokens, true
		}
		lval = jsonpathSymType{}
		posBeforeScan = p.scanner.Pos()
		p.scanner.Scan(&lval)
		if lval.id == 0 {
			return p.scanner.In()[startPos:posBeforeScan], tokens, (lval.id == 0)
		}
		lval.pos -= startPos
		tokens = append(tokens, lval)
	}
}

func (p *Parser) parse(
	query string, tokens []jsonpathSymType, nakedIntType *types.T,
) (*jsonpath.Jsonpath, error) {
	p.lexer.init(query, tokens, nakedIntType, &p.parserImpl)
	defer p.lexer.cleanup()
	if p.parserImpl.Parse(&p.lexer) != 0 {
		if p.lexer.lastError == nil {
			p.lexer.Error("syntax error")
		}
		err := p.lexer.lastError
		return nil, err
	}
	return p.lexer.expr, nil
}

func (p *Parser) Parse(jsonpath string, nakedIntType *types.T) (*jsonpath.Jsonpath, error) {
	p.scanner.Init(jsonpath)
	defer p.scanner.Cleanup()

	query, tokens, done := p.scan()
	stmt, err := p.parse(query, tokens, nakedIntType)
	if err != nil {
		return nil, err
	}
	if !done {
		return nil, errors.AssertionFailedf("invalid jsonpath query: %s", jsonpath)
	}
	return stmt, nil
}

func Parse(jsonpath string) (*jsonpath.Jsonpath, error) {
	var p Parser
	return p.Parse(jsonpath, defaultNakedIntType)
}

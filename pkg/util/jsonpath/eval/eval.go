// Copyright 2025 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package eval

import (
	"errors"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/jsonpath"
	"github.com/cockroachdb/cockroach/pkg/util/jsonpath/parser"
)

var UnimplementedError = errors.New("unimplemented")
var UnknownTypeError = errors.New("unknown type")

func JsonpathQuery(target tree.DJSON, path tree.DJsonpath) ([]tree.DJSON, error) {
	jp, err := parser.Parse(string(path))
	if err != nil {
		return []tree.DJSON{}, err
	}

	if len(jp.Query.Accessors) == 0 {
		panic("at least one accessor is guaranteed")
	}
	_, ok := jp.Query.Accessors[0].(jsonpath.Root)
	if !ok {
		panic("the first accessor is the root")
	}

	res := []tree.DJSON{target}
	for _, accessor := range jp.Query.Accessors[1:] {
		switch a := accessor.(type) {
		case jsonpath.Key:
			var cur []tree.DJSON
			for _, r := range res {
				c, err := r.JSON.FetchValKey(a.Key)
				if err != nil {
					return []tree.DJSON{}, err
				}
				if c == nil {
					if jp.Strict {
						return []tree.DJSON{}, pgerror.Newf(pgcode.KeyNotInJSON,
							"JSON object does not contain key %q", a.Key)
					}
					continue
				}
				cur = append(cur, tree.DJSON{JSON: c})
			}
			res = cur
		case jsonpath.Wildcard:
			var cur []tree.DJSON
			for _, r := range res {
				paths, err := json.AllPathsWithDepth(r.JSON, 1)
				if err != nil {
					return []tree.DJSON{}, err
				}
				for _, path := range paths {
					stripped, err := path.FetchValIdx(0)
					if err != nil {
						return []tree.DJSON{}, err
					}
					cur = append(cur, tree.DJSON{JSON: stripped})
				}
			}
			res = cur
		default:
			return []tree.DJSON{}, UnknownTypeError
		}
	}
	return res, nil
}

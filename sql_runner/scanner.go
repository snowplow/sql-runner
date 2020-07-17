//
// Copyright (c) 2015-2018 Snowplow Analytics Ltd. All rights reserved.
//
// This program is licensed to you under the Apache License Version 2.0,
// and you may not use this file except in compliance with the Apache License Version 2.0.
// You may obtain a copy of the Apache License Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0.
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the Apache License Version 2.0 is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the Apache License Version 2.0 for the specific language governing permissions and limitations there under.
//
package main

import (
	"github.com/go-pg/pg/orm"
)

type Results struct {
	results  [][]string
	columns  []string
	elements int
	rows     int
}

var _ orm.HooklessModel = (*Results)(nil)

func (results *Results) Init() error {
	results.elements = 0
	results.rows = 0

	if s := results; len(s.results) >= 0 {
		results.results = (s.results)[:0]
	}

	if s := results; len(s.columns) >= 0 {
		results.columns = (s.columns)[:0]
	}
	return nil
}

func (results *Results) NewModel() orm.ColumnScanner {
	return results
}

func (Results) AddModel(_ orm.ColumnScanner) error {
	return nil
}

func (results *Results) ScanColumn(colIdx int, colName string, b []byte) error {
	curRow := len(results.results) - 1

	if colIdx == 0 {
		results.results = append(results.results, []string{})
		curRow = len(results.results) - 1
		results.rows += 1
	}

	if curRow == 0 {
		results.columns = append(results.columns, colName)
	}

	results.elements += 1
	results.results[curRow] = append(results.results[curRow], string(b))
	return nil
}

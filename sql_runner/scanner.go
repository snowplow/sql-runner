//
// Copyright (c) 2015-2022 Snowplow Analytics Ltd. All rights reserved.
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
	"github.com/go-pg/pg/v10/orm"
	"github.com/go-pg/pg/v10/types"
)

// Results information.
type Results struct {
	results  [][]string
	columns  []string
	elements int
	rows     int
}

var _ orm.HooklessModel = (*Results)(nil)

// Init initializes Results.
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

// NextColumnScanner returns a ColumnScanner that is used to scan columns.
func (results *Results) NextColumnScanner() orm.ColumnScanner {
	return results
}

// AddColumnScanner adds the ColumnScanner to the model.
func (Results) AddColumnScanner(_ orm.ColumnScanner) error {
	return nil
}

// ScanColumn implements ColumnScanner interface.
func (results *Results) ScanColumn(col types.ColumnInfo, rd types.Reader, n int) error {
	tmp, err := rd.ReadFullTemp()
	if err != nil {
		return err
	}

	curRow := len(results.results) - 1

	if col.Index == 0 {
		results.results = append(results.results, []string{})
		curRow = len(results.results) - 1
		results.rows++
	}

	if curRow == 0 {
		results.columns = append(results.columns, col.Name)
	}

	results.elements++
	results.results[curRow] = append(results.results[curRow], string(tmp))
	return nil
}

package main

import (
	"github.com/go-pg/pg/orm"
)

type Results struct {
	results [][]string
	columns []string
	elements int
	rows int
}

var _ orm.HooklessModel = (*Results)(nil)
//var _ types.ValueAppender = (*Results)(nil)

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
	curRow := len(results.results)-1

	if colIdx == 0 {
		results.results = append(results.results, []string{})
		curRow = len(results.results)-1
		results.rows += 1
	}


	if curRow == 0 {
		results.columns = append(results.columns, colName)
	}

	results.elements += 1
	results.results[curRow] = append(results.results[curRow], string(b))
	return nil
}

/*
func (results Results) AppendValue(dst []byte, quote int) []byte {
	if len(results.results) <= 0 {
		return dst
	}

	for _, s := range results.results {
		dst = types.AppendString(dst, s, 1)
		dst = append(dst, ',')
	}
	dst = dst[:len(dst)-1]
	return dst
}*/
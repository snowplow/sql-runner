package orm

import (
	"reflect"

	"gopkg.in/pg.v5/internal"
)

type sliceModel struct {
	slice reflect.Value
	scan  func(reflect.Value, []byte) error
}

var _ Model = (*sliceModel)(nil)

func (m *sliceModel) Reset() error {
	if m.slice.IsValid() && m.slice.Len() > 0 {
		m.slice.Set(m.slice.Slice(0, 0))
	}
	return nil
}

func (m *sliceModel) NewModel() ColumnScanner {
	return m
}

func (sliceModel) AddModel(_ ColumnScanner) error {
	return nil
}

func (sliceModel) AfterQuery(_ DB) error {
	return nil
}

func (sliceModel) AfterSelect(_ DB) error {
	return nil
}

func (sliceModel) BeforeInsert(_ DB) error {
	return nil
}

func (sliceModel) AfterInsert(_ DB) error {
	return nil
}

func (sliceModel) BeforeUpdate(_ DB) error {
	return nil
}

func (sliceModel) AfterUpdate(_ DB) error {
	return nil
}

func (sliceModel) BeforeDelete(_ DB) error {
	return nil
}

func (sliceModel) AfterDelete(_ DB) error {
	return nil
}

func (m *sliceModel) ScanColumn(colIdx int, _ string, b []byte) error {
	v := internal.SliceNextElem(m.slice)
	return m.scan(v, b)
}

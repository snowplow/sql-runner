package orm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type structTableModel struct {
	table *Table
	joins []join

	root  reflect.Value
	index []int

	strct reflect.Value
}

var _ tableModel = (*structTableModel)(nil)

func newStructTableModel(v interface{}) (*structTableModel, error) {
	switch v := v.(type) {
	case *structTableModel:
		return v, nil
	case reflect.Value:
		return newStructTableModelValue(v)
	default:
		return newStructTableModelValue(reflect.ValueOf(v))
	}
}

func newStructTableModelValue(v reflect.Value) (*structTableModel, error) {
	if !v.IsValid() {
		return nil, errors.New("pg: Model(nil)")
	}
	v = reflect.Indirect(v)

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("pg: Model(unsupported %s)", v.Type())
	}

	return &structTableModel{
		table: Tables.Get(v.Type()),
		root:  v,
		strct: v,
	}, nil
}

func (structTableModel) useQueryOne() bool {
	return true
}

func (m *structTableModel) Table() *Table {
	return m.table
}

func (m *structTableModel) AppendParam(dst []byte, name string) ([]byte, bool) {
	if field, ok := m.table.FieldsMap[name]; ok {
		dst = field.AppendValue(dst, m.strct, 1)
		return dst, true
	}

	if method, ok := m.table.Methods[name]; ok {
		dst = method.AppendValue(dst, m.strct.Addr(), 1)
		return dst, true
	}

	switch name {
	case "TableAlias":
		dst = append(dst, m.table.Alias...)
		return dst, true
	}

	return dst, false
}

func (m *structTableModel) Root() reflect.Value {
	return m.root
}

func (m *structTableModel) Index() []int {
	return m.index
}

func (m *structTableModel) Bind(bind reflect.Value) {
	m.strct = bind.Field(m.index[len(m.index)-1])
}

func (m *structTableModel) Value() reflect.Value {
	return m.strct
}

func (m *structTableModel) bindChildren() {
	for i := range m.joins {
		j := &m.joins[i]
		switch j.Rel.Type {
		case HasOneRelation, BelongsToRelation:
			j.JoinModel.Bind(m.strct)
		}
	}
}

func (m *structTableModel) NewModel() ColumnScanner {
	m.bindChildren()
	return m
}

func (m *structTableModel) AddModel(_ ColumnScanner) error {
	return nil
}

func (m *structTableModel) AfterQuery(db DB) error {
	if !m.table.Has(AfterQueryHookFlag) {
		return nil
	}
	return callAfterQueryHook(m.strct.Addr(), db)
}

func (m *structTableModel) AfterSelect(db DB) error {
	if !m.table.Has(AfterSelectHookFlag) {
		return nil
	}
	return callAfterSelectHook(m.strct.Addr(), db)
}

func (m *structTableModel) BeforeCreate(db DB) error {
	if !m.table.Has(BeforeCreateHookFlag) {
		return nil
	}
	return callBeforeCreateHook(m.strct.Addr(), db)
}

func (m *structTableModel) AfterCreate(db DB) error {
	if !m.table.Has(AfterCreateHookFlag) {
		return nil
	}
	return callAfterCreateHook(m.strct.Addr(), db)
}

func (m *structTableModel) ScanColumn(colIdx int, colName string, b []byte) error {
	ok, err := m.scanColumn(colIdx, colName, b)
	if ok {
		return err
	}
	return fmt.Errorf("pg: can't find column %s in model %s", colName, m.table.Type.Name())
}

func (m *structTableModel) scanColumn(colIdx int, colName string, b []byte) (bool, error) {
	joinName, fieldName := splitColumn(colName)
	if joinName != "" {
		if join := m.GetJoin(joinName); join != nil {
			return join.JoinModel.scanColumn(colIdx, fieldName, b)
		}
		if m.table.ModelName == joinName {
			return m.scanColumn(colIdx, fieldName, b)
		}
	}

	field, ok := m.table.FieldsMap[colName]
	if ok {
		if m.strct.Kind() == reflect.Interface {
			m.strct = m.strct.Elem()
		}
		if m.strct.Kind() == reflect.Ptr {
			if m.strct.IsNil() {
				m.strct.Set(reflect.New(m.strct.Type().Elem()))
				m.strct = m.strct.Elem()
				m.bindChildren()
			} else {
				m.strct = m.strct.Elem()
			}
		}
		return true, field.ScanValue(m.strct, b)
	}

	return false, nil
}

func (m *structTableModel) GetJoin(name string) *join {
	for i := range m.joins {
		j := &m.joins[i]
		if j.Rel.Field.GoName == name || j.Rel.Field.SQLName == name {
			return j
		}
	}
	return nil
}

func (m *structTableModel) GetJoins() []join {
	return m.joins
}

func (m *structTableModel) AddJoin(j join) *join {
	m.joins = append(m.joins, j)
	return &m.joins[len(m.joins)-1]
}

func (m *structTableModel) Join(name string, apply func(*Query) *Query) *join {
	return m.join(m.Value(), name, apply)
}

func (m *structTableModel) join(bind reflect.Value, name string, apply func(*Query) *Query) *join {
	path := strings.Split(name, ".")
	index := make([]int, 0, len(path))

	currJoin := join{
		BaseModel: m,
		JoinModel: m,
	}
	var lastJoin *join
	var hasColumnName bool

	for _, name := range path {
		rel, ok := currJoin.JoinModel.Table().Relations[name]
		if !ok {
			hasColumnName = true
			break
		}
		currJoin.Rel = rel
		index = append(index, rel.Field.Index...)

		if j := currJoin.JoinModel.GetJoin(name); j != nil {
			currJoin.BaseModel = j.BaseModel
			currJoin.JoinModel = j.JoinModel
			lastJoin = j
		} else {
			model, err := newTableModelIndex(bind, index, rel.JoinTable)
			if err != nil {
				return nil
			}

			currJoin.Parent = lastJoin
			currJoin.BaseModel = currJoin.JoinModel
			currJoin.JoinModel = model
			lastJoin = currJoin.BaseModel.AddJoin(currJoin)
		}
	}

	// No joins with such name.
	if lastJoin == nil {
		return nil
	}
	if apply != nil {
		lastJoin.ApplyQuery = apply
	}

	if hasColumnName {
		column := path[len(path)-1]
		lastJoin.Columns = append(lastJoin.Columns, column)
	}

	return lastJoin
}

func splitColumn(s string) (string, string) {
	ind := strings.Index(s, "__")
	if ind == -1 {
		return "", s
	}
	return s[:ind], s[ind+2:]
}

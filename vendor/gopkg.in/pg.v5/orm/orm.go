package orm

import "gopkg.in/pg.v5/types"

// ColumnScanner is used to scan column values.
type ColumnScanner interface {
	// Scan assigns a column value from a row.
	//
	// An error should be returned if the value can not be stored
	// without loss of information.
	ScanColumn(colIdx int, colName string, b []byte) error
}

type QueryAppender interface {
	AppendQuery(dst []byte, params ...interface{}) ([]byte, error)
}

type QueryFormatter interface {
	FormatQuery(dst []byte, query string, params ...interface{}) []byte
}

// DB is a common interface for pg.DB and pg.Tx types.
type DB interface {
	Model(model ...interface{}) *Query
	Select(model interface{}) error
	Insert(model ...interface{}) error
	Update(model interface{}) error
	Delete(model interface{}) error

	Exec(query interface{}, params ...interface{}) (*types.Result, error)
	ExecOne(query interface{}, params ...interface{}) (*types.Result, error)
	Query(coll, query interface{}, params ...interface{}) (*types.Result, error)
	QueryOne(model, query interface{}, params ...interface{}) (*types.Result, error)

	QueryFormatter
}

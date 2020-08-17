package mysql

import (
	"database/sql"
)

//返回一行数据,然后再用Scan方法返回数据，注意，这个地方先Prepare安全性似乎更高一点
func (m *MySqlStruct) QueryRow(query string, args ...interface{}) *sql.Row {
	stmt, _ := m.DB.Prepare(query)
	defer stmt.Close()
	return stmt.QueryRow(args...)
}

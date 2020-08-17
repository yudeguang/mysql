//对MySql进行简单封装,简化查询,增,删,改等操作,同时也支持原生MySql所有操作
package mysql

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"runtime"
)

//DB大写导出,故一些诸如事务等高级用法可直接调用*sql.DB实现
type MySqlStruct struct {
	DB        *sqlx.DB
	goVersion string //golang版本
}

/*
连接数据库:NewMysql函数的通俗化别名
Sql, err := Open(`root:123456@tcp(localhost:3306)/test?charset=utf8`)
*/
func Open(conn string) (*MySqlStruct, error) {
	var err error
	var m MySqlStruct
	m.goVersion = runtime.Version()

	m.DB, err = sqlx.Open(`mysql`, conn)
	if err != nil {
		return nil, err
	}
	err = m.DB.Ping()
	return &m, err
}

//最大连接数
func (m *MySqlStruct) SetMaxOpenConns(n int) {
	m.DB.SetMaxOpenConns(n)
}

//最大空闲数
func (m *MySqlStruct) SetMaxIdleConns(n int) {
	m.DB.SetMaxIdleConns(n)
}

//关闭数据库
// err:=Close()
func (m *MySqlStruct) Close() error {
	return m.DB.Close()
}

//对MySql进行简单封装,简化查询,增,删,改等操作,同时也支持原生MySql所有操作
package mysql

import (
	"database/sql"
	"fmt"
	"strings"
)

/*
插入操作:
原始SQL:
  insert into calss(name,age,sex)values('andy', 18 ,'男')
Insert:
  _,err:=Insert(`insert into calss(name,age,sex) values(?,?,?)`, "andy", 18, "男")
*/
func (m *MySqlStruct) Insert(query string, args ...interface{}) (sql.Result, error) {
	if !(strings.HasPrefix(strings.TrimSpace(strings.ToLower(query)), `insert into `) ||
		strings.HasPrefix(strings.TrimSpace(strings.ToLower(query)), `insert ignore into `)) {
		return nil, fmt.Errorf(fmt.Sprintf("there is a err in your query: %v .", query))
	}
	return m.Exec(query, args...)
}

/*
插入操作:
InsertEasy为针对插入操作时，Insert方法更为简洁的替代方案，尤其是在列数较多的情况下尤为实用.
原始SQL:
  insert into calss(name,age,sex)values('andy', 18 ,'男')
Insert:
  _,err:=Insert(`insert into calss(name,age,sex) values(?,?,?)`, "andy", 18, "男")
InsertEasy:
	_,err:=InsertEasy(`class`, `name,age,sex`, "andy", 18, "男")

参数说明:
intoTableName==>待插入的表名
columns==>插入时涉及的相关列
args==>插入时涉及相关列的值
*/
func (m *MySqlStruct) InsertEasy(intoTableName, columns string, args ...interface{}) (sql.Result, error) {
	return m.Insert(m.getInsertQuery(intoTableName, columns), args...)
}

/*
插入操作:
InsertIgnoreEasy为针对插入操作时，对应InsertEasy方法只插入不重复的数据
*/
func (m *MySqlStruct) InsertIgnoreEasy(intoTableName, columns string, args ...interface{}) (sql.Result, error) {
	return m.Insert(strings.Replace(m.getInsertQuery(intoTableName, columns), `insert into `, `insert ignore into `, 1), args...)
}

/*
插入操作:
InsertFromStruct为针对插入操作时，待插入的数据已经在某一结构体中.
原始SQL:
  insert into calss(name,age,sex)values('andy', 18 ,'男')
Insert:
  _,err:=Insert(`insert into calss(name,age,sex) values(?,?,?)`, "andy", 18, "男")
InsertEasy:
	_,err:=InsertEasy(`class`, `name,age,sex`, "andy", 18, "男")
InsertFromStruct:
	type Student struct {
		Name string
		age  int
		sex  string
	}
	andyyu := Student{"andy", 18, "男"}
	_,err:=InsertFromStruct("class",andyyu)

参数说明:
intoTableName==>待插入的表名
fromStructName==>插入时所用的数据(数据放在结构体中)
*/
func (m *MySqlStruct) InsertFromStruct(intoTableName string, fromStructName interface{}) (result sql.Result, err error) {
	fieldNames, args, err := StructToSlice(fromStructName)
	if err != nil {
		return nil, err
	}
	//返回
	return m.Insert(m.getInsertQuery(intoTableName, strings.Join(fieldNames, ",")), args...)
}

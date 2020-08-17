//对MySql进行简单封装,简化查询,增,删,改等操作,同时也支持原生MySql所有操作
package mysql

import (
	"database/sql"
)

/*
原则上支持insert,update,delete等所有操作,除此之外，支持创建表，修改表等操作
原始SQL:
	CREATE TABLE `class` (
	`id`  int NULL ,
	`name`  varchar(255) NULL ,
	`age`  int NULL ,
	`sex`  varchar(255) NULL ,
	PRIMARY KEY (`id`)
	);
Exec:
	query:=`CREATE TABLE class (
	id  int NULL ,
	name  varchar(255) NULL ,
	age  int NULL ,
	sex  varchar(255) NULL ,
	PRIMARY KEY (id)
	)`
  _,err:=Exec(query)
*/
func (m *MySqlStruct) Exec(query string, args ...interface{}) (sql.Result, error) {
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	return stmt.Exec(args...)
}

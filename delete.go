//对MySql进行简单封装,简化查询,增,删,改等操作,同时也支持原生MySql所有操作
package mysql

import (
	"database/sql"
	"fmt"
	"strings"
)

/*
删除操作
原始SQL:
  delete from  calss where name='andy' and  age=18 and sex='男')
Delete:
   _,err:=Delete(`delete from calss where name=? and age=? and sex="男"`, "andy", 18, "男")
*/
func (m *MySqlStruct) Delete(query string, args ...interface{}) (sql.Result, error) {
	if !strings.HasPrefix(strings.TrimSpace(strings.ToLower(query)), `delete `) {
		return nil, fmt.Errorf(fmt.Sprintf("there is a err in your query: %v .", query))
	}
	return m.Exec(query, args...)
}

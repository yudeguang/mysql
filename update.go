//对MySql进行简单封装,简化查询,增,删,改等操作,同时也支持原生MySql所有操作
package mysql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

/*
更新操作:
原始SQL:
  update calss set name='andyyu',age=20 where id=6
Update:
  _,err:=Update(`update calss set name=?,age=?`, "andyyu", 20, 6)
*/
func (m *MySqlStruct) Update(query string, args ...interface{}) (sql.Result, error) {
	if !strings.HasPrefix(strings.TrimSpace(strings.ToLower(query)), `update `) {
		return nil, fmt.Errorf(fmt.Sprintf("there is a err in your query: %v .", query))
	}
	return m.Exec(query, args...)
}

/*
更新数据，根据表主键等条件更新单条或者多条数据
	type Student struct {
		ID   int
		Name string
		Age  int
		Sex  string
	}
	andyyu := Student{6,"余德光", 18, "男"}
	_, err = sql.UpdateFromStruct("class", andyyu, "id")

toBeUpdateTableName==>指待更新的表名
fromStructName==>更新所用的数据源(在一结构体中的数据)
conditionColumns==>属于fromStructName中的一个或者多个列名，若列不不属于fromStructName会被忽略
*/
func (m *MySqlStruct) UpdateFromStruct(toBeUpdateTableName string, fromStructName interface{}, conditionColumns ...string) (sql.Result, error) {
	toBeUpdateColumns, args, err := StructToSlice(fromStructName)
	if err != nil {
		return nil, err
	}
	//把主键对应的值添加入args
	refValue := reflect.ValueOf(fromStructName)
	fieldNum := refValue.NumField()
	for i := 0; i < fieldNum; i++ {
		for _, conditionColumn := range conditionColumns {
			if strings.ToLower(refValue.Type().Field(i).Name) == strings.ToLower(conditionColumn) {
				args = append(args, value2Interface(refValue.Field(i)))
				break
			}
		}
	}
	return m.Update(m.getUpdateQuery(toBeUpdateTableName, toBeUpdateColumns, conditionColumns...), args...)
}

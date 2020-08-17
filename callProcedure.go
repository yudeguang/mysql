//对MySql进行简单封装,简化查询,增,删,改等操作,同时也支持原生MySql所有操作
package mysql

import (
	"fmt"
	"strings"
)

/*
执行存储过程,传入一个切片的指针，结果被返回到该切片中；
该函数支持Prepare
结构体字段顺序必须与查询列顺序一致.
函数只返回第一个结果集，如果想处理多个结果集，调用MySqlStruct.Mymysqldb自行处理.
原始SQL:
  call pro_get_students("男",12)
CallProcedure:
  type student struct {
		Id   int
		Name string
		Age  int
  }
  var result []student
  err = CallProcedure(&result, `call pro_get_students(?,?)`, "男", 12)
*/
func (m *MySqlStruct) CallProcedure(intoResultPtr interface{}, query string, args ...interface{}) error {
	if !m.canRunInCurrentEnvironment() {
		return fmt.Errorf(fmt.Sprintf("CallProcedure can not run in %v,this func can run at least go1.9", m.goVersion))
	}
	if !strings.HasPrefix(strings.TrimSpace(strings.ToLower(query)), `call `) {
		return fmt.Errorf(fmt.Sprintf("there is a err in your query: %v .", query))
	}
	return m.Select(intoResultPtr, query, args...)
}

/*
执行存储过程，只返回第一个结果集，如果想处理多个结果集，调用MySqlStruct.Mymysqldb自行处理.
数据库中有NUll不会报错，NUll值被替换成文本"NULL".
该函数支持Prepare.
取第i行用:row:=result[i]
取某行的第i个字段用v=:row[i]
原始SQL:
  call pro_get_students("男",12)
CallProcedureSafeSlice:
  result,err:=CallProcedureSafeSlice(`pro_get_students(?,?)`, "男",12)
*/
func (m *MySqlStruct) CallProcedureSafeSlice(query string, args ...interface{}) ([][]string, error) {
	if !m.canRunInCurrentEnvironment() {
		return nil, fmt.Errorf(fmt.Sprintf("CallProcedureSafeSlice can not run in %v,this func can run at least go1.9", m.goVersion))
	}
	if !strings.HasPrefix(strings.TrimSpace(strings.ToLower(query)), `call `) {
		return nil, fmt.Errorf(fmt.Sprintf("there is a err in your query: %v .", query))
	}
	return m.SelectSafeSlice(query, args...)
}

/*
执行存储过程，只返回第一个结果集，如果想处理多个结果集，调用MySqlStruct.Mymysqldb自行处理.
数据库中有NUll不会报错，NUll值被替换成文本"NULL".
暂不支持Prepare,需自行拼接SQL.
取第i行用:row:=result[i]
取某行的名称为id的字段用v:=row["id"],注意字段名称大小写敏感，需与数据库查询字段名称保持一致
原始SQL:
  call pro_get_students("男",12)
CallProcedureSafeMap:
  result,err:=CallProcedureSafeMap(`pro_get_students(?,?)`, "男",12)
*/
func (m *MySqlStruct) CallProcedureSafeMap(query string, args ...interface{}) ([]map[string]string, error) {
	if !m.canRunInCurrentEnvironment() {
		return nil, fmt.Errorf(fmt.Sprintf("CallProcedureSafeMap can not run in %v,this func can run at least go1.9", m.goVersion))
	}
	if !strings.HasPrefix(strings.TrimSpace(strings.ToLower(query)), `call `) {
		return nil, fmt.Errorf(fmt.Sprintf("there is a err in your query: %v .", query))
	}
	return m.SelectSafeMap(query, args...)
}

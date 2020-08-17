//对MySql进行简单封装,简化查询,增,删,改等操作,同时也支持原生MySql所有操作
package mysql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

/*
查询,返回*sql.Rows,一般用于对性能要求较高，后续自行遍历处理sql.Rows的情况.
原始SQL:
  select id,name,age from calss where sex='男' and age>12;
SelectRows:
	rows,err:=SelectRows(`select id,name,age from calss where sex=? and age>?`,"男",12)
后续使用方法示例:
type student struct {
	id   int
	age  int
	name string
	sex  string
}
func getStudents(rows *sql.Rows) (result []student, err error) {
	defer rows.Close()
	for rows.Next() {
		var oneRow student
		err := rows.Scan(&oneRow.id, &oneRow.age, &oneRow.name, &oneRow.sex)
		if err != nil {
			return nil, err
		}
		result = append(result, oneRow)
	}
	return
}
*/
func (m *MySqlStruct) SelectRows(query string, args ...interface{}) (*sql.Rows, error) {
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	return stmt.Query(args...)
}

/*
查询,传入一个切片的指针，结果被返回到该切片中；
原始SQL:
  select id,name,age from calss where sex='男' and age>12;
Select:
	type student struct {
		Id   int
		Name string
		Age  int
	}
	var result []student
	err = Select(&result, `select id,name,age from calss where sex=? and age>?`, "男", 12)
*/
func (m *MySqlStruct) Select(intoResultPtr interface{}, query string, args ...interface{}) error {
	//先判断传入的数据是否是指针,now the value shoule be: *[]interface{},top kind is a ptr
	refValue := reflect.ValueOf(intoResultPtr)
	if refValue.Kind() != reflect.Ptr { //&& refValue.IsNil()
		return fmt.Errorf("the first argument resultPtr must be a pointer,not a value.")
	}
	//再判断下一级是否是切片,now dirValue shoule be: []interface{},top kind is a slice
	dirValue := reflect.Indirect(refValue)
	if dirValue.Kind() != reflect.Slice {
		return fmt.Errorf("the first argument resultPtr must be a slice.")
	}
	//判断切片是否为空
	if l := dirValue.Len(); l != 0 {
		return fmt.Errorf(fmt.Sprintf("the first argument resultPtr has %v records,and it's must be empty.", l))
	}
	/*
		再判断切片元素类型，只支持int,int8...等基础类型以及结构体.
		其它类型则诸如:Uintptr,Complex64,Complex128,Array,Chan,
		Func,Interface,Map,Ptr,Slice,UnsafePointer,直接报错返回.
	*/
	itemNum := 1
	IsBaseType := false
	structElem := reflect.Value{}
	structObject := reflect.New(dirValue.Type().Elem())
	arrayObject := reflect.MakeSlice(dirValue.Type(), 0, 0)
	switch dirValue.Type().Elem().Kind() {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64,
		reflect.String:
		IsBaseType = true
	case reflect.Struct:
		//判断结构体中字段的数字母是否是大写，因为反射只在大写情况才起作用
		structElem = structObject.Elem()
		itemNum = structElem.NumField()
		for i := 0; i < itemNum; i++ {
			if !structElem.Field(i).CanSet() {
				fieldName := structElem.Type().Field(i).Name
				intoResultPtrName := dirValue.Type().Elem().Name()
				return fmt.Errorf(fmt.Sprintf("the field name %v.%v should be %v.%v,because the first letter is capitalized can be exported in reflect.",
					intoResultPtrName, fieldName, intoResultPtrName, strings.Title(fieldName)))
			}
		}
	default:
		return fmt.Errorf("the first argument resultPtr is not a support type.")
	}
	rows, err := m.SelectRows(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	//判断两方元素数量是否一样多
	if itemNum != len(columns) {
		return fmt.Errorf(fmt.Sprintf("intoResultPtr fileds num %v doesn't mutch columns num %v from database.", itemNum, len(columns)))
	}
	for rows.Next() { //rows.NextResultSet()
		oneRowPtr := make([]interface{}, itemNum)
		//实例化oneRowPtr
		i := 0 //出错时，要取列名
		for i = 0; i < itemNum; i++ {
			if IsBaseType {
				oneRowPtr[i] = structObject.Interface()
			} else {
				oneRowPtr[i] = structElem.Field(i).Addr().Interface()
			}
		}

		//Scan到oneRowPtr，也就意味着Scan到structElem
		err = rows.Scan(oneRowPtr...)
		if err != nil {
			fieldName := structElem.Type().Field(i - 1).Name
			columnName := columns[i-1]
			return fmt.Errorf(fmt.Sprintf("intoResultPtr %vth fileds %v doesn't mutch database %vth column %v or %v.", i, fieldName, i, columnName, err))
		}
		arrayObject = reflect.Append(arrayObject, structObject.Elem())
	}
	dirValue.Set(arrayObject)
	return nil
}

//旧方法，因历史原因保留,被替换为SelectSafeSlice
func (m *MySqlStruct) Query(query string, args ...interface{}) ([][]string, error) {
	return m.SelectSafeSlice(query, args...)
}

/*
查询,返回[][]string类型数据，即最终所有数据都被转化为string存储.
数据库中有NUll不会报错，NUll值被替换成文本"NULL".
取第i行用:row:=result[i]
取某行的第i个字段用v=:row[i]
原始SQL:
  select id,name,age from calss where sex='男' and age>12;
SelectSafeSlice:
  result,err:=SelectSafeSlice(`select id,name,age from calss where sex=? and age>?`,"男",12)
*/
func (m *MySqlStruct) SelectSafeSlice(query string, args ...interface{}) ([][]string, error) {
	rows, err := m.SelectRows(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	//获取列数量
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	lenColumns := len(columns)
	var result [][]string
	for rows.Next() {
		oneRow := make([]string, lenColumns)
		oneRowHasNULL := make([]sql.RawBytes, lenColumns)
		oneRowPtr := make([]interface{}, lenColumns)
		//实例化oneRowPtr，oneRowPtr中元素的值存放在oneRowHasNULL中
		for i := 0; i < lenColumns; i++ {
			oneRowPtr[i] = &oneRowHasNULL[i]
		}
		//Scan到oneRowPtr，也就意味着Scan到oneRowHasNULL，此时oneRowHasNULL已有数据
		err = rows.Scan(oneRowPtr...)
		if err != nil {
			return nil, err
		}
		//处理NULL值
		for i, v := range oneRowHasNULL {
			if v == nil {
				oneRow[i] = "NULL"
			} else {
				oneRow[i] = string(oneRowHasNULL[i])
			}
		}
		result = append(result, oneRow)
	}
	return result, nil
}

/*
查询,返回[]map[string]string类型数据，即最终所有数据都被转化为string存储.
数据库中有NUll不会报错，NUll值被替换成文本"NULL".
取第i行用:row:=result[i]
取某行的名称为id的字段用v:=row["id"],注意字段名称大小写敏感，需与数据库查询字段名称保持一致
原始SQL:
  select id,name,age from calss where sex='男' and age>12;
SelectSafeMap:
  result,err:=SelectSafeMap(`select id,name,age from calss where sex=? and age>?`,"男",12)
*/
func (m *MySqlStruct) SelectSafeMap(query string, args ...interface{}) ([]map[string]string, error) {
	rows, err := m.SelectRows(query, args...)
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result := make([]map[string]string, 0)
	lenColumns := len(columns)
	for rows.Next() {
		oneRow := make([]string, lenColumns)
		oneRowHasNULL := make([]sql.RawBytes, lenColumns)
		oneRowPtr := make([]interface{}, lenColumns)
		for i := 0; i < lenColumns; i++ {
			oneRowPtr[i] = &oneRowHasNULL[i]
		}
		err = rows.Scan(oneRowPtr...)
		if err != nil {
			return nil, err
		}
		//处理NULL值
		for i, v := range oneRowHasNULL {
			if v == nil {
				oneRow[i] = "NULL"
			} else {
				oneRow[i] = string(oneRowHasNULL[i])
			}
		}
		tempMap := make(map[string]string, lenColumns)
		for i := 0; i < lenColumns; i++ {
			tempMap[columns[i]] = oneRow[i]
		}
		result = append(result, tempMap)
	}
	return result, nil
}

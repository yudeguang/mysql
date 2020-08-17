//对MySql进行简单封装,简化查询,增,删,改等操作,同时也支持原生MySql所有操作
package mysql

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

//检测传入的SQL代码中的参数是否安全,裸写代码时,为防注入有时会用上
func (m *MySqlStruct) IsSqlParameterSafe(s string) bool {
	s = strings.ToLower(s)
	str := `(?:')|(?:--)|(/\\*(?:.|[\\n\\r])*?\\*/)|(\b(select|update|and|or|delete|insert|trancate|char|chr|into|substr|ascii|declare|exec|count|master|into|drop|execute)\b)`
	re, err := regexp.Compile(str)
	if err != nil {
		panic(err)
	}
	return !re.MatchString(s)
}

/*
拼接拼接InsertQuery:
"class","age,name" ==>  insert into class(age,name) values(?,?)
*/
func (m *MySqlStruct) getInsertQuery(tableName, toBeInsertColumns string) string {
	//返回N个以逗号分隔的?占位符字符串 5 ==> ?,?,?,?,?
	var nQuestions = func(n int) string {
		if n < 1 {
			panic("n must > 0.")
		}
		var buffer bytes.Buffer
		//前面若干条中间要加","
		for i := 0; i < n-1; i++ {
			buffer.WriteString("?,")
		}
		//最后一次不加","
		buffer.WriteString("?")
		return buffer.String()
	}
	return `insert into ` + tableName + `(` + toBeInsertColumns + `) values(` + nQuestions(len(strings.Split(toBeInsertColumns, `,`))) + `)`
}

/*
拼接UpdateQuery:
必须要有where 条件，以免功能被滥用
"class","age,name","id" ==>  update class set age=?,name=? where id=?
*/
func (m *MySqlStruct) getUpdateQuery(tableName string, toBeUpdateColumns []string, conditionColumns ...string) string {
	sqlText := "update " + tableName + " set "
	for _, v := range toBeUpdateColumns {
		sqlText = sqlText + v + "=?,"
	}
	sqlText = strings.TrimSuffix(sqlText, ",") + " where "

	for _, v := range conditionColumns {

		sqlText = sqlText + v + "=? and "
	}
	return strings.TrimSuffix(sqlText, "and ")
}

/*
判断函数是否可运行在当前环境
*/
func (m *MySqlStruct) canRunInCurrentEnvironment() bool {
	//返回左侧N个字符
	var left = func(s string, n int) string {
		if n <= 0 || s == "" {
			return ""
		}
		runes := []rune(s)
		if len(runes) <= n {
			return s
		}
		return string(runes[0:n])
	}
	if left(m.goVersion, 1) > "1" ||
		left(m.goVersion, 3) == "1.9" ||
		left(m.goVersion, 3) == "1.10" || //go 2 之前估计最多也就1.10 1.11 1.12 1.13....
		left(m.goVersion, 3) == "1.11" ||
		left(m.goVersion, 3) == "1.12" ||
		left(m.goVersion, 3) == "1.13" ||
		left(m.goVersion, 3) == "1.14" ||
		left(m.goVersion, 3) == "1.15" ||
		left(m.goVersion, 3) == "1.16" ||
		left(m.goVersion, 3) == "1.17" ||
		left(m.goVersion, 3) == "1.18" ||
		left(m.goVersion, 3) == "1.19" {
		return true
	}
	return false
}

//value转化为Interface
func value2Interface(fieldValue reflect.Value) interface{} {
	fieldType := fieldValue.Type()
	k := fieldType.Kind()
	switch k {
	case reflect.Bool:
		return fieldValue.Bool()
		//Int()返回的是int64,而不是int
	case reflect.Int:
		return int(fieldValue.Int())
	case reflect.Int8:
		return int8(fieldValue.Int())
	case reflect.Int16:
		return int16(fieldValue.Int())
	case reflect.Int32:
		return int32(fieldValue.Int())
	case reflect.Int64:
		return int64(fieldValue.Int())
		//Uint()返回的是uint64,而不是uint
	case reflect.Uint:
		return uint(fieldValue.Uint())
	case reflect.Uint8:
		return uint8(fieldValue.Uint())
	case reflect.Uint16:
		return uint16(fieldValue.Uint())
	case reflect.Uint32:
		return uint32(fieldValue.Uint())
	case reflect.Uint64:
		return uint64(fieldValue.Uint())
		//Float()返回的是float64
	case reflect.Float32:
		return float32(fieldValue.Float())
	case reflect.Float64:
		return float64(fieldValue.Float())
		//Complex()返回的是complex128
	case reflect.Complex64:
		return complex64(fieldValue.Complex())
	case reflect.Complex128:
		return complex128(fieldValue.Complex())
		//case reflect.Array:
		// case reflect.Chan:
		// case reflect.Func:
		// case reflect.Interface:
		// case reflect.Map:
	case reflect.Ptr:
		return fieldValue.Pointer()
	// case reflect.Slice:
	case reflect.String:
		return fieldValue.String()
	// case reflect.Struct:
	default:
		return fieldValue.Interface()
	}

}

//把结构体中的数据转到切片中
func StructToSlice(fromStructName interface{}) (fieldNames []string, slice []interface{}, err error) {
	//确保fromStructName是结构体
	refValue := reflect.ValueOf(fromStructName)
	if refValue.Kind() != reflect.Struct {
		err = fmt.Errorf("the argument fromStructName must be a Struct.")
		return
	}
	//反射获得所有列名及值
	fieldNum := refValue.NumField()
	slice = make([]interface{}, 0, 2)
	fieldNames = make([]string, 0, fieldNum)
	for i := 0; i < fieldNum; i++ {
		fieldNames = append(fieldNames, refValue.Type().Field(i).Name)
		slice = append(slice, value2Interface(refValue.Field(i)))
	}

	return
}

/*
用于在使用mysql.Select函数时,拼接前半部分SQL语句(select ,,,,,, ),在结构体字段较多时尤为实用.
例:
type Student struct {
		ID   int
		Name string
		Age  int
		Sex  string
	}
selectSqlText:=mysql.GetSelectSqlTextFrom(Student{})
selectSqlText==> select ID,Name,Age,Sex

参数说明:
fromStructName==>待查询字段所在的结构体
*/
//
func GetSelectSqlTextFrom(structName interface{}) string {
	//确保fromStructName是结构体
	refValue := reflect.ValueOf(structName)
	if refValue.Kind() != reflect.Struct {
		return "the argument fromStructName must be a Struct."
	}
	sqlText := "select "
	//反射获得所有列名
	fieldNum := refValue.NumField()
	for i := 0; i < fieldNum; i++ {
		sqlText = sqlText + refValue.Type().Field(i).Name + ","
	}
	return strings.TrimSuffix(sqlText, ",") + " "
}

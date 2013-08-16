package sqlutil

import (
	"bytes"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"time"
)

var ColumnNameToFieldName func(string) string = SnakeToUpperCamel

// 查询单条数据
func One(out interface{}, rows *sql.Rows) (err error) {
	defer rows.Close()

	panicMsg := errors.New("expected pointer to struct *struct")
	if x := reflect.TypeOf(out).Kind(); x != reflect.Ptr {
		panic(panicMsg)
	}

	rowValue := reflect.ValueOf(out)
	if x := reflect.Indirect(rowValue).Type().Kind(); x != reflect.Struct {
		panic(panicMsg)
	}

	for rows.Next() {
		err = ScanRow(rowValue, rows)
		break
	}
	return
}

// 查询多条数据
func All(out interface{}, rows *sql.Rows) (err error) {
	defer rows.Close()

	panicMsg := errors.New("expected pointer to struct slice *[]struct")
	if x := reflect.TypeOf(out).Kind(); x != reflect.Ptr {
		panic(panicMsg)
	}

	sliceValue := reflect.Indirect(reflect.ValueOf(out))
	if x := sliceValue.Kind(); x != reflect.Slice {
		panic(panicMsg)
	}

	sliceType := sliceValue.Type().Elem()
	if x := sliceType.Kind(); x != reflect.Struct {
		panic(panicMsg)
	}

	for rows.Next() {
		rowValue := reflect.New(sliceType)
		err = ScanRow(rowValue, rows)
		if err != nil {
			return
		}

		if rowValue.Type().Kind() == reflect.Ptr {
			rowValue = rowValue.Elem()
		}
		sliceValue.Set(reflect.Append(sliceValue, rowValue))
	}
	return
}

// 扫描一行
func ScanRow(rowValue reflect.Value, rows *sql.Rows) (err error) {
	cols, _ := rows.Columns()
	containers := make([]interface{}, 0, len(cols))

	for i := 0; i < cap(containers); i++ {
		var v interface{}
		containers = append(containers, &v)
	}

	err = rows.Scan(containers...)
	if err != nil {
		return
	}

	for i, v := range containers {
		value := reflect.Indirect(reflect.ValueOf(v))
		key := cols[i]
		if !value.Elem().IsValid() {
			continue
		}
		field := rowValue.Elem().FieldByName(ColumnNameToFieldName(key))
		if field.IsValid() {
			SetModelValue(value, field)
		}
	}
	return
}

// 通过反射设置单个元素的值
func SetModelValue(driverValue, fieldValue reflect.Value) error {
	fieldType := fieldValue.Type()
	switch fieldType.Kind() {
	case reflect.Bool:
		if driverValue.Elem().Int() != 0 {
			fieldValue.Set(reflect.ValueOf(true))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fieldValue.SetInt(driverValue.Elem().Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch driverValue.Elem().Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldValue.SetUint(uint64(driverValue.Elem().Int()))
		default:
			fieldValue.SetUint(driverValue.Elem().Uint())
		}
	case reflect.Float32, reflect.Float64:
		fieldValue.SetFloat(driverValue.Elem().Float())
	case reflect.String:
		fieldValue.SetString(string(driverValue.Elem().Bytes()))
	case reflect.Slice:
		if reflect.TypeOf(driverValue.Interface()).Elem().Kind() == reflect.Uint8 {
			fieldValue.SetBytes(driverValue.Elem().Bytes())
		}
	case reflect.Struct:
		if fieldType == reflect.TypeOf(time.Time{}) {
			fieldValue.Set(driverValue.Elem())
		}
	}
	return nil
}

// 下划线命名方式转换为大写的驼峰
func SnakeToUpperCamel(s string) string {
	buf := bytes.NewBuffer(nil)
	for _, v := range strings.Split(s, "_") {
		if len(v) > 0 {
			buf.WriteString(strings.ToUpper(v[:1]))
			buf.WriteString(v[1:])
		}
	}
	return buf.String()
}

// 大写的驼峰转换为下划线命名方式
func UpperCamelToSnake(s string) string {
	buf := bytes.NewBuffer(nil)
	for i, v := range s {
		if i > 0 && v >= 'A' && v <= 'Z' {
			buf.WriteRune('_')
		}
		buf.WriteRune(v)
	}
	return strings.ToLower(buf.String())
}

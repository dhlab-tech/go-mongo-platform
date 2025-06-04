package inmemory

import (
	"fmt"
	"reflect"
	"strings"
)

func getStringFieldValueByName(in any, field string) string {
	p := reflect.ValueOf(in)
	if p.Kind() == reflect.Ptr {
		return getStringFieldValueByName(p.Elem().Interface(), field)
	}
	_field := strings.Split(field, "+")
	if len(_field) > 1 {
		return getStringFieldValueByName(p.FieldByName(_field[0]).Interface(), strings.Join(_field[1:], "+"))
	}
	if p.FieldByName(field).Kind() == reflect.Ptr {
		p = p.FieldByName(field).Elem()
	} else {
		p = p.FieldByName(field)
	}
	switch p.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", p.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", p.Uint())
	case reflect.Bool:
		return fmt.Sprintf("%v", p.Bool())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", p.Float())
	case reflect.String:
		return p.String()
	default:
		fmt.Println("HEHE100", in, p, field)
		if !p.IsNil() {
			return fmt.Sprintf("%v", p.Interface())
		}
	}
	return ""
}

func getStringFieldValuesByName(in any, fields []string) string {
	var res []string
	for _, f := range fields {
		res = append(res, getStringFieldValueByName(in, f))
	}
	return strings.Join(res, "")
}

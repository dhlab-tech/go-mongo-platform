package inmemory

import (
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
	if p.Kind() == reflect.String {
		return p.Interface().(string)
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

package inmemory

import (
	"fmt"
	"reflect"
	"strings"
)

// func addStringFieldValueByName(in any, field string) string {
// 	p := reflect.ValueOf(in)
// 	if p.Kind() == reflect.Ptr {
// 		return addStringFieldValueByName(p.Elem().Interface(), field)
// 	}
// 	_field := strings.Split(field, "+")
// 	if len(_field) > 1 {
// 		return addStringFieldValueByName(p.FieldByName(_field[0]).Interface(), strings.Join(_field[1:], "+"))
// 	}
// 	if p.FieldByName(field).Kind() == reflect.Ptr {
// 		if !p.FieldByName(field).IsNil() {
// 			p = p.FieldByName(field).Elem()
// 		} else {
// 			_t := reflect.TypeOf(p.FieldByName(field).Interface())
// 			p = reflect.New(_t.Elem()).Elem()
// 		}
// 	} else {
// 		p = p.FieldByName(field)
// 	}
// 	switch p.Kind() {
// 	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
// 		return fmt.Sprintf("%d", p.Int())
// 	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
// 		return fmt.Sprintf("%d", p.Uint())
// 	case reflect.Bool:
// 		return fmt.Sprintf("%v", p.Bool())
// 	case reflect.Float32, reflect.Float64:
// 		return fmt.Sprintf("%f", p.Float())
// 	case reflect.String:
// 		return p.String()
// 	default:
// 		if !p.IsNil() {
// 			return fmt.Sprintf("%v", p.Interface())
// 		}
// 	}
// 	return ""
// }

// func addStringFieldValuesByName(in any, fields []string) string {
// 	var res []string
// 	for _, f := range fields {
// 		res = append(res, addStringFieldValueByName(in, f))
// 	}
// 	return strings.Join(res, "")
// }

func updateStringFieldValueByName(in any, field string) *string {
	p := reflect.ValueOf(in)
	if p.Kind() == reflect.Ptr {
		return updateStringFieldValueByName(p.Elem().Interface(), field)
	}
	_field := strings.Split(field, "+")
	if len(_field) > 1 {
		return updateStringFieldValueByName(p.FieldByName(_field[0]).Interface(), strings.Join(_field[1:], "+"))
	}
	if p.FieldByName(field).Kind() == reflect.Ptr {
		if !p.FieldByName(field).IsNil() {
			p = p.FieldByName(field).Elem()
		} else {
			return nil
		}
	} else {
		p = p.FieldByName(field)
	}
	switch p.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return ptr(fmt.Sprintf("%d", p.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return ptr(fmt.Sprintf("%d", p.Uint()))
	case reflect.Bool:
		return ptr(fmt.Sprintf("%v", p.Bool()))
	case reflect.Float32, reflect.Float64:
		return ptr(fmt.Sprintf("%f", p.Float()))
	case reflect.String:
		return ptr(p.String())
	default:
		if !p.IsNil() {
			return ptr(fmt.Sprintf("%v", p.Interface()))
		}
	}
	return nil
}

func updateStringFieldValuesByName(in any, fields []string) *string {
	var res []string
	for _, f := range fields {
		r := updateStringFieldValueByName(in, f)
		if r != nil {
			res = append(res, *r)
		}
	}
	return ptr(strings.Join(res, ""))
}

func ptr(in string) *string {
	_in := in
	return &_in
}

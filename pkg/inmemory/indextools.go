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

func updateStringFieldValueByName(in any, field string) (r *string) {
	defer func() {
		if r != nil && *r == "" {
			r = nil
		}
	}()
	p := reflect.ValueOf(in)
	if p.Kind() == reflect.Ptr {
		if p.IsNil() {
			return nil
		}
		elem := p.Elem()
		if !elem.IsValid() {
			return nil
		}
		return updateStringFieldValueByName(elem.Interface(), field)
	}
	_field := strings.Split(field, "+")
	if len(_field) > 1 {
		fieldVal := p.FieldByName(_field[0])
		if !fieldVal.IsValid() {
			return nil
		}
		if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
			return nil
		}
		if fieldVal.Kind() == reflect.Ptr {
			fieldVal = fieldVal.Elem()
		}
		if !fieldVal.IsValid() {
			return nil
		}
		return updateStringFieldValueByName(fieldVal.Interface(), strings.Join(_field[1:], "+"))
	}
	fieldVal := p.FieldByName(field)
	if !fieldVal.IsValid() {
		return nil
	}
	if fieldVal.Kind() == reflect.Ptr {
		if fieldVal.IsNil() {
			return nil
		}
		p = fieldVal.Elem()
		if !p.IsValid() {
			return nil
		}
	} else {
		p = fieldVal
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
		if p.IsValid() && !p.IsNil() {
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
	if len(res) == 0 {
		return nil
	}
	return ptr(strings.Join(res, ""))
}

func ptr(in string) *string {
	_in := in
	return &_in
}

func _updateStringFieldValueByName(in any, field string) (r []string) {
	p := reflect.ValueOf(in)
	if p.Kind() == reflect.Ptr {
		if p.IsNil() {
			return nil
		}
		elem := p.Elem()
		if !elem.IsValid() {
			return nil
		}
		return _updateStringFieldValueByName(elem.Interface(), field)
	}
	_field := strings.Split(field, "+")
	if len(_field) > 1 {
		fieldVal := p.FieldByName(_field[0])
		if !fieldVal.IsValid() {
			return nil
		}
		if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
			return nil
		}
		if fieldVal.Kind() == reflect.Ptr {
			fieldVal = fieldVal.Elem()
		}
		if !fieldVal.IsValid() {
			return nil
		}
		return _updateStringFieldValueByName(fieldVal.Interface(), strings.Join(_field[1:], "+"))
	}
	fieldVal := p.FieldByName(field)
	if !fieldVal.IsValid() {
		return nil
	}
	if fieldVal.Kind() == reflect.Ptr {
		if fieldVal.IsNil() {
			return nil
		}
		p = fieldVal.Elem()
		if !p.IsValid() {
			return nil
		}
	} else {
		p = fieldVal
	}
	switch p.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return append(r, fmt.Sprintf("%d", p.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return append(r, fmt.Sprintf("%d", p.Uint()))
	case reflect.Bool:
		return append(r, fmt.Sprintf("%v", p.Bool()))
	case reflect.Float32, reflect.Float64:
		return append(r, fmt.Sprintf("%f", p.Float()))
	case reflect.String:
		return append(r, p.String())
	case reflect.Slice:
		for i := 0; i < p.Len(); i++ {
			switch p.Index(i).Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				r = append(r, fmt.Sprintf("%d", p.Index(i).Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				r = append(r, fmt.Sprintf("%d", p.Index(i).Uint()))
			case reflect.Bool:
				r = append(r, fmt.Sprintf("%v", p.Index(i).Bool()))
			case reflect.Float32, reflect.Float64:
				r = append(r, fmt.Sprintf("%f", p.Index(i).Float()))
			case reflect.String:
				r = append(r, p.Index(i).String())
			}
		}
		return
	default:
		if p.IsValid() && !p.IsNil() {
			return append(r, fmt.Sprintf("%v", p.Interface()))
		}
	}
	return nil
}

// Cannot create composite index on slices
// On a slice, you can only create an index on a single field
func _updateStringFieldValuesByName(in any, fields []string) []string {
	if len(fields) == 1 {
		return _updateStringFieldValueByName(in, fields[0])
	}
	var res []string
	for _, f := range fields {
		r := _updateStringFieldValueByName(in, f)
		if len(r) == 1 {
			res = append(res, r[0])
		}
	}
	if len(res) == 0 {
		return nil
	}
	return append([]string{}, strings.Join(res, ""))
}

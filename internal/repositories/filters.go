package repositories

import (
	"reflect"

	"github.com/breakfront-planner/auth-service/internal/autherrors"
)

// FilterField is a struct with name and values of one filter field.
type FilterField struct {
	FilterName string
	DBName     string
	Value      any
}

// ParseFilter get fields from filter and return db name and value.
func ParseFilter(filter any) ([]FilterField, error) {
	v := reflect.ValueOf(filter)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	var fields []FilterField

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if field.Kind() != reflect.Ptr {
			return nil, autherrors.ErrNoPtrsFilterFields
		}

		columnName := fieldType.Tag.Get("db")
		if columnName == "" || field.IsNil() {
			continue
		}

		fields = append(fields, FilterField{
			FilterName: fieldType.Name,
			DBName:     columnName,
			Value:      field.Elem().Interface(),
		})
	}

	if len(fields) == 0 {
		return nil, autherrors.ErrEmptyFilter
	}

	return fields, nil
}

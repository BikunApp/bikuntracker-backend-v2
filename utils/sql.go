package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
)

func GetPartialUpdateSQL[T any](tableName string, data T, whereData *models.WhereData) (sql string, params []interface{}, err error) {
	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	if v.Kind() != reflect.Struct {
		sql = ""
		err = fmt.Errorf("Given data is not a type struct")
		return
	}

	sql = fmt.Sprintf("UPDATE %s SET ", tableName)

	params = make([]interface{}, 0)
	setClauses := []string{}
	setCounter := 1

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.IsNil() {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", strings.Split(fieldType.Tag.Get("json"), ",")[0], setCounter))
			if field.Kind() == reflect.Pointer {
				params = append(params, field.Elem().Interface())
			} else {
				params = append(params, field.Interface())
			}
			setCounter++
		}
	}

	sql += strings.Join(setClauses, ", ") + fmt.Sprintf(" WHERE %s = $%d", whereData.FieldName, setCounter)
	params = append(params, whereData.Value)

	return
}

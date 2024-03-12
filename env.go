//go:build !go1.18

package envvar

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
)

var (
	TagEnvVar    = "env"
	TagDefault   = "default"
	TagRequired  = "required"
	TagConverter = "converter"
)

type ConverterFunc func(string) (any, error)

var convFunctions map[string]ConverterFunc

func Converter(key string, convFun ConverterFunc) {
	if convFunctions == nil {
		convFunctions = make(map[string]ConverterFunc)
	}
	convFunctions[key] = convFun
}

func ClearConverters() {
	convFunctions = nil
}

func Load(v any) error {
	t := reflect.TypeOf(v)
	var el reflect.Value
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		el = reflect.ValueOf(v).Elem()
	} else {
		return errors.New("must pass pointer to a struct")
	}

	return decodePtrStruct(t, el)
}

func decodePtrStruct(structType reflect.Type, structVal reflect.Value) error {
	num := structVal.NumField()
	for i := 0; i < num; i++ {
		field := structType.Field(i)
		fieldType := field.Type
		fieldVal := structVal.Field(i)
		varName := field.Tag.Get(TagEnvVar)
		defaultValue := field.Tag.Get(TagDefault)
		requiredVal := field.Tag.Get(TagRequired)
		converter := field.Tag.Get(TagConverter)

		if fieldType.Kind() == reflect.Ptr && reflect.Struct == fieldType.Elem().Kind() && !fieldVal.IsNil() {
			err := decodePtrStruct(fieldType.Elem(), fieldVal.Elem())
			if err != nil {
				return err
			}
		}

		// new value from envvar/default
		var newValue string
		if defaultValue != "" && fieldVal.IsZero() {
			newValue = defaultValue
		}
		if varName != "" {
			if ev := os.Getenv(varName); ev != "" {
				newValue = ev
			}
		}

		// required validation
		if strings.ToLower(requiredVal) == "true" && fieldVal.IsZero() && newValue == "" {
			return errors.New(fmt.Sprintf("%v is required", getFieldName(structType, field)))
		}

		if newValue == "" {
			continue
		}

		if converter != "" {
			if convFunc, ok := convFunctions[converter]; ok {
				converted, err := convFunc(newValue)
				if err != nil {
					return errors.New(fmt.Sprintf("%v - %v", getFieldName(structType, field), err.Error()))
				}
				if fieldType.String() != reflect.ValueOf(converted).Type().String() {
					return errors.New(fmt.Sprintf("%v - incorrect converter", getFieldName(structType, field)))
				}
				fieldVal.Set(reflect.ValueOf(converted))
			} else {
				return errors.New(fmt.Sprintf("%v - converter %v does not exist",
					getFieldName(structType, field), converter))
			}
		} else {
			switch fieldType.String() {
			case "string":
				fieldVal.SetString(newValue)
			case "*string":
				fieldVal.Set(reflect.ValueOf(ptr(newValue)))
			default:
				return errors.New(fmt.Sprintf("%v - must define converter function for non string or "+
					"*string field", getFieldName(structType, field)))
			}
		}

	}
	return nil
}

func getFieldName(structType reflect.Type, field reflect.StructField) string {
	return fmt.Sprintf("%v.%v", structType.Name(), field.Name)
}

func ptr[T any](v T) *T {
	return &v
}

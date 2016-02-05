package filters

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type FieldError struct {
	FieldPath   string `json:"field_path"`
	FieldName   string `json:"field_name"`
	ErrorString string `json:"error_string"`
}

// SanitizeStruct will sanitize all fields of a struct (o must be a pointer to this struct)
// return an array of string, one string per error
func SanitizeStruct(o interface{}) []FieldError {
	errors := []FieldError{}
	sanitizeStruct(o, "", &errors)
	return errors
}

func sanitizeStruct(o interface{}, path string, errors *[]FieldError) {
	st := reflect.TypeOf(o)
	vt := reflect.ValueOf(o)

	if st.Kind() == reflect.Ptr {
		st = st.Elem()
		vt = vt.Elem()
	}

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		value := vt.Field(i)

		n_path := ""
		if path != "" {
			n_path = path + "." + field.Name
		} else {
			n_path = field.Name
		}

		fmt.Println("path", n_path)

		if field.Type.Kind() == reflect.Struct {
			sanitizeStruct(value.Interface(), n_path, errors)
		} else {
			sanitizeField(field, value, n_path, errors)
		}
	}
}

// DefaultStruct will set all defaults value to fields that have a default set
// this function is to call before filling the struct with form values
func DefaultStruct(o interface{}) {
	st := reflect.TypeOf(o)
	vt := reflect.ValueOf(o)

	if st.Kind() == reflect.Ptr {
		st = st.Elem()
		vt = vt.Elem()
	}

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		value := vt.Field(i)
		fmt.Println("defaulting field", i, ":", field.Name)
		if field.Type.Kind() == reflect.Struct {
			DefaultStruct(value.Interface())
		} else {
			setFieldToDefault(field, value)
		}
	}
}

func sanitizeField(field reflect.StructField, value reflect.Value, path string, errors *[]FieldError) {
	sanitizeFieldMin(field, value, path, errors)
	sanitizeFieldMax(field, value, path, errors)
}

// setFieldToDefault if there is a default value
func setFieldToDefault(field reflect.StructField, value reflect.Value) {
	s_default := field.Tag.Get("default")

	if len(s_default) == 0 {
		return
	}

	switch field.Type.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		def, _ := strconv.ParseInt(s_default, 10, 64)
		value.SetInt(def)
	case reflect.Float32, reflect.Float64:
		def, _ := strconv.ParseFloat(s_default, 64)
		value.SetFloat(def)
	case reflect.String:
		value.SetString(s_default)
	default:
		log.Println("SetFieldDefault on type", field.Type.Name(), "not implemented")
	}
}

func setFieldError(field reflect.StructField, value reflect.Value, path string, errors *[]FieldError) {
	log.Println("setting", field.Name, "to default", field.Tag.Get("default"))
	setFieldToDefault(field, value)
	s_error := field.Tag.Get("error")
	if len(s_error) > 0 {
		*errors = append(*errors, FieldError{
			FieldPath:   path,
			FieldName:   field.Name,
			ErrorString: s_error,
		})
	}
}

// sanitizeFieldMin check if field value is bellow minimum. If true, true is returned
func sanitizeFieldMin(field reflect.StructField, value reflect.Value, path string, errors *[]FieldError) bool {
	s_min := field.Tag.Get("min")

	if len(s_min) == 0 {
		return false
	}

	switch field.Type.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		min, _ := strconv.ParseInt(s_min, 10, 64)
		if value.Int() < min {
			setFieldError(field, value, path, errors)
			return true
		}
	case reflect.Float32, reflect.Float64:
		min, _ := strconv.ParseFloat(s_min, 64)
		if value.Float() < min {
			setFieldError(field, value, path, errors)
			return true
		}
	case reflect.String:
		min, _ := strconv.Atoi(s_min)
		if len(value.String()) < min {
			setFieldError(field, value, path, errors)
			return true
		}
	default:
		log.Println("SanitizeFieldMin on type", field.Type.Name(), "not implemented")
		return true
	}
	return false
}

// sanitizeFieldMax check if field value is above maxium. If true, true is returned
func sanitizeFieldMax(field reflect.StructField, value reflect.Value, path string, errors *[]FieldError) bool {
	s_max := field.Tag.Get("max")

	if len(s_max) == 0 {
		return false
	}

	switch field.Type.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		max, _ := strconv.ParseInt(s_max, 10, 64)
		if value.Int() > max {
			setFieldError(field, value, path, errors)
			return true
		}
	case reflect.Float32, reflect.Float64:
		max, _ := strconv.ParseFloat(s_max, 64)
		if value.Float() > max {
			setFieldError(field, value, path, errors)
			return true
		}
	case reflect.String:
		max, _ := strconv.Atoi(s_max)
		if len(value.String()) > max {
			setFieldError(field, value, path, errors)
			return true
		}
	default:
		log.Println("SanitizeFieldMax on type", field.Type.Name(), "not implemented")
		return true
	}
	return false
}

// sanitizeFieldEnum check if field value is above maxium. If true, true is returned
func sanitizeFieldEnum(field reflect.StructField, value reflect.Value, path string, errors *[]FieldError) bool {
	s_enum := field.Tag.Get("enum")

	if len(s_enum) == 0 {
		return false
	}

	s_enums := strings.Split(s_enum, ",")

	switch field.Type.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val := value.Int()
		for _, s_enum := range s_enums {
			v, _ := strconv.ParseInt(s_enum, 10, 64)
			if val == v {
				return false
			}
		}
		setFieldError(field, value, path, errors)
		return true
	case reflect.Float32, reflect.Float64:
		val := value.Float()
		for _, s_enum := range s_enums {
			v, _ := strconv.ParseFloat(s_enum, 64)
			if val == v {
				return false
			}
		}
		setFieldError(field, value, path, errors)
		return true
	case reflect.String:
		val := value.String()
		for _, s_enum := range s_enums {
			if val == s_enum {
				return false
			}
		}
		setFieldError(field, value, path, errors)
		return true
	default:
		log.Println("SanitizeFieldEnum on type", field.Type.Name(), "not implemented")
		return true
	}
	return false
}

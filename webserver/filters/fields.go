package filters

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
)

// SanitizeStruct will sanitize all fields of a struct (o must be a pointer to this struct)
func SanitizeStruct(o interface{}) {
	fmt.Println("sanitisz : ", o)
	st := reflect.TypeOf(o)
	vt := reflect.ValueOf(o)

	if st.Kind() == reflect.Ptr {
		st = st.Elem()
		vt = vt.Elem()
	}

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		value := vt.Field(i)
		fmt.Println("field", i, ":", field.Name)
		if field.Type.Kind() == reflect.Struct {
			fmt.Println("sub sanitize : ", field.Name)
			SanitizeStruct(value.Interface())
		} else {
			sanitizeField(field, value)
		}
	}
}

// DefaultStruct will set all defaults value to fields that have a default set
// this function is to call before filling the struct with form values
func DefaultStruct(o interface{}) {
	st := reflect.TypeOf(o).Elem()
	vt := reflect.ValueOf(o).Elem()
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		value := vt.Field(i)
		fmt.Println("field", i, ":", field.Name)
		setFieldToDefault(field, value)
	}
}

func sanitizeField(field reflect.StructField, value reflect.Value) {
	sanitizeFieldMin(field, value)
	sanitizeFieldMax(field, value)
}

// setFieldToDefault if there is a default value
func setFieldToDefault(field reflect.StructField, value reflect.Value) {
	s_default := field.Tag.Get("default")

	log.Println("setting", field.Name, "to default", s_default)

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
	default:
		log.Println("SetFieldDefault on type", field.Type.Name(), "not implemented")
	}
}

// sanitizeFieldMin check if field value is bellow minimum. If true, true is returned
func sanitizeFieldMin(field reflect.StructField, value reflect.Value) bool {
	s_min := field.Tag.Get("min")

	if len(s_min) == 0 {
		return false
	}

	switch field.Type.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		min, _ := strconv.ParseInt(s_min, 10, 64)
		if value.Int() < min {
			setFieldToDefault(field, value)
			return true
		}
	case reflect.Float32, reflect.Float64:
		min, _ := strconv.ParseFloat(s_min, 64)
		if value.Float() < min {
			setFieldToDefault(field, value)
			return true
		}
	default:
		log.Println("SanitizeFieldMin on type", field.Type.Name(), "not implemented")
		return true
	}
	return false
}

// sanitizeFieldMax check if field value is above maxium. If true, true is returned
func sanitizeFieldMax(field reflect.StructField, value reflect.Value) bool {
	s_max := field.Tag.Get("max")

	if len(s_max) == 0 {
		return false
	}

	switch field.Type.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		max, _ := strconv.ParseInt(s_max, 10, 64)
		if value.Int() > max {
			setFieldToDefault(field, value)
			return true
		}
	case reflect.Float32, reflect.Float64:
		max, _ := strconv.ParseFloat(s_max, 64)
		if value.Float() > max {
			setFieldToDefault(field, value)
			return true
		}
	default:
		log.Println("SanitizeFieldMax on type", field.Type.Name(), "not implemented")
		return true
	}
	return false
}

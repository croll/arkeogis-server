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

		name := field.Tag.Get("json")
		if len(name) == 0 {
			name = field.Name
		}

		n_path := ""
		if path != "" {
			n_path = path + "." + name
		} else {
			n_path = name
		}

		fmt.Println("path", n_path)

		if field.Type.Kind() == reflect.Struct {
			sanitizeStruct(value.Interface(), n_path, errors)
		} else {
			sanitizeField(field, value, n_path, name, errors)
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

type Tag struct {
	Name     string
	Value    string
	TagError *Tag
}

func decodeTags(field reflect.StructField) (tags []Tag) {
	// Inspired from go /src/reflect/type.go
	tag := field.Tag

	tags = []Tag{}

	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		value, err := strconv.Unquote(qvalue)
		if err != nil {
			break
		}

		if name == "error" {
			errtag := Tag{
				Name:  name,
				Value: value,
			}
			for _, t := range tags {
				if t.TagError == nil {
					t.TagError = &errtag
				}
			}
		} else {
			tags = append(tags, Tag{
				Name:  name,
				Value: value,
			})
		}
	}
	return tags
}

func sanitizeField(field reflect.StructField, value reflect.Value, path string, fieldname string, errors *[]FieldError) {
	tags := decodeTags(field)
	for _, tag := range tags {
		switch tag.Name {
		case "min":
			sanitizeFieldMin(field, tag, value, path, fieldname, errors)
		case "max":
			sanitizeFieldMax(field, tag, value, path, fieldname, errors)
		case "enum":
			sanitizeFieldEnum(field, tag, value, path, fieldname, errors)
		}
	}
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

func setFieldError(field reflect.StructField, tag Tag, value reflect.Value, path string, fieldname string, errors *[]FieldError) {
	log.Println("setting", field.Name, "to default", field.Tag.Get("default"))
	setFieldToDefault(field, value)

	if tag.TagError != nil && len(tag.TagError.Value) > 0 {
		*errors = append(*errors, FieldError{
			FieldPath:   path,
			FieldName:   fieldname,
			ErrorString: tag.TagError.Value,
		})
	}
}

// sanitizeFieldMin check if field value is bellow minimum. If true, true is returned
func sanitizeFieldMin(field reflect.StructField, tag Tag, value reflect.Value, path string, fieldname string, errors *[]FieldError) bool {
	if len(tag.Value) == 0 {
		return false
	}

	switch field.Type.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		min, _ := strconv.ParseInt(tag.Value, 10, 64)
		if value.Int() < min {
			setFieldError(field, tag, value, path, fieldname, errors)
			return true
		}
	case reflect.Float32, reflect.Float64:
		min, _ := strconv.ParseFloat(tag.Value, 64)
		if value.Float() < min {
			setFieldError(field, tag, value, path, fieldname, errors)
			return true
		}
	case reflect.String:
		min, _ := strconv.Atoi(tag.Value)
		if len(value.String()) < min {
			setFieldError(field, tag, value, path, fieldname, errors)
			return true
		}
	default:
		log.Println("SanitizeFieldMin on type", field.Type.Name(), "not implemented")
		return true
	}
	return false
}

// sanitizeFieldMax check if field value is above maxium. If true, true is returned
func sanitizeFieldMax(field reflect.StructField, tag Tag, value reflect.Value, path string, fieldname string, errors *[]FieldError) bool {
	if len(tag.Value) == 0 {
		return false
	}

	switch field.Type.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		max, _ := strconv.ParseInt(tag.Value, 10, 64)
		if value.Int() > max {
			setFieldError(field, tag, value, path, fieldname, errors)
			return true
		}
	case reflect.Float32, reflect.Float64:
		max, _ := strconv.ParseFloat(tag.Value, 64)
		if value.Float() > max {
			setFieldError(field, tag, value, path, fieldname, errors)
			return true
		}
	case reflect.String:
		max, _ := strconv.Atoi(tag.Value)
		if len(value.String()) > max {
			setFieldError(field, tag, value, path, fieldname, errors)
			return true
		}
	default:
		log.Println("SanitizeFieldMax on type", field.Type.Name(), "not implemented")
		return true
	}
	return false
}

// sanitizeFieldEnum check if field value is above maxium. If true, true is returned
func sanitizeFieldEnum(field reflect.StructField, tag Tag, value reflect.Value, path string, fieldname string, errors *[]FieldError) bool {
	if len(tag.Value) == 0 {
		return false
	}

	s_enums := strings.Split(tag.Value, ",")

	switch field.Type.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val := value.Int()
		for _, s_enum := range s_enums {
			v, _ := strconv.ParseInt(s_enum, 10, 64)
			if val == v {
				return false
			}
		}
		setFieldError(field, tag, value, path, fieldname, errors)
		return true
	case reflect.Float32, reflect.Float64:
		val := value.Float()
		for _, s_enum := range s_enums {
			v, _ := strconv.ParseFloat(s_enum, 64)
			if val == v {
				return false
			}
		}
		setFieldError(field, tag, value, path, fieldname, errors)
		return true
	case reflect.String:
		val := value.String()
		for _, s_enum := range s_enums {
			if val == s_enum {
				return false
			}
		}
		setFieldError(field, tag, value, path, fieldname, errors)
		return true
	default:
		log.Println("SanitizeFieldEnum on type", field.Type.Name(), "not implemented")
		return true
	}
	return false
}

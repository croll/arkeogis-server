package sanitizer

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
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
	st := reflect.TypeOf(o)
	vt := reflect.ValueOf(o)
	sanitizeStruct(st, vt, nil, "", "", &errors)
	return errors
}

func sanitizeStruct(st reflect.Type, vt reflect.Value, field *reflect.StructField, path string, name string, errors *[]FieldError) {
	fmt.Println("path : ", path, "name: ", name)
	switch st.Kind() {
	case reflect.Ptr:
		st = st.Elem()
		vt = vt.Elem()
		sanitizeStruct(st, vt, nil, path, name, errors)
	case reflect.Struct:
		for i := 0; i < st.NumField(); i++ {
			field := st.Field(i)
			if field.Name[:1] >= "a" && field.Name[:1] <= "z" {
				continue
			}
			value := vt.Field(i)

			name := field.Tag.Get("json")
			if len(name) == 0 && !field.Anonymous {
				name = field.Name
			}

			n_path := ""
			if field.Anonymous {
				n_path = path
			} else {
				if path != "" {
					n_path = path + "." + name
				} else {
					n_path = name
				}
			}

			fmt.Println("path", n_path)

			sanitizeStruct(field.Type, value, &field, n_path, name, errors)
		}
	case reflect.Array, reflect.Map, reflect.Slice:
		log.Println("type ", st.Kind, " unsupported, todo !")
	case reflect.Invalid, reflect.Chan, reflect.Interface, reflect.UnsafePointer:
		log.Println("type ", st.Kind, " unsupported")
	default:
		sanitizeField(*field, vt, path, name, errors)
	}
}

// DefaultStruct will set all defaults value to fields that have a default set
// this function is to call before filling the struct with form values
func DefaultStruct(o interface{}) {
	st := reflect.TypeOf(o)
	vt := reflect.ValueOf(o)
	defaultStruct(st, vt, nil)
}

func defaultStruct(st reflect.Type, vt reflect.Value, field *reflect.StructField) {
	switch st.Kind() {
	case reflect.Ptr:
		st = st.Elem()
		vt = vt.Elem()
		defaultStruct(st, vt, nil)
	case reflect.Struct:
		for i := 0; i < st.NumField(); i++ {
			field := st.Field(i)
			if field.Name[:1] >= "a" && field.Name[:1] <= "z" {
				continue
			}
			value := vt.Field(i)
			defaultStruct(field.Type, value, &field)
		}
	case reflect.Array, reflect.Map, reflect.Slice:
		log.Println("type ", st.Kind, " unsupported, todo !")
	case reflect.Invalid, reflect.Chan, reflect.Interface, reflect.UnsafePointer:
		log.Println("type ", st.Kind, " unsupported")
	default:
		setFieldToDefault(*field, vt)
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
			for i, _ := range tags {
				if tags[i].TagError == nil {
					tags[i].TagError = &errtag
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
		case "set":
			sanitizeFieldSet(field, tag, value, path, fieldname, errors)
		case "regexp":
			sanitizeFieldRegexp(field, tag, value, path, fieldname, errors)
		case "email":
			sanitizeFieldEmail(field, tag, value, path, fieldname, errors)
		}
	}
}

// setFieldToDefault if there is a default value
func setFieldToDefault(field reflect.StructField, value reflect.Value) {
	s_default := field.Tag.Get("default")

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

// sanitizeFieldEnum check if field value is in an enum of values (separated by a comma)
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

// sanitizeFieldSet check that multiple values separated by a comma are in a set of values that are also separated by a comma
func sanitizeFieldSet(field reflect.StructField, tag Tag, value reflect.Value, path string, fieldname string, errors *[]FieldError) bool {
	if len(tag.Value) == 0 {
		return false
	}

	s_sets := strings.Split(tag.Value, ",")

	switch field.Type.Kind() {
	case reflect.String:
		vals := strings.Split(value.String(), ",")
		for _, val := range vals {
			found := false
			for _, s_set := range s_sets {
				if val == s_set {
					found = true
					break
				}
			}
			if !found {
				setFieldError(field, tag, value, path, fieldname, errors)
				return true
			}
		}
	default:
		log.Println("SanitizeFieldSet on type", field.Type.Name(), "not implemented")
		return true
	}
	return false
}

// sanitizeFieldRegexp a value using regular expression
func sanitizeFieldRegexp(field reflect.StructField, tag Tag, value reflect.Value, path string, fieldname string, errors *[]FieldError) bool {
	if len(tag.Value) == 0 {
		return false
	}

	switch field.Type.Kind() {
	case reflect.String:
		matched, err := regexp.MatchString(tag.Value, value.String())
		log.Println("reg: ", tag.Value, "on :", value.String, "matched:", matched, "err:", err)
		if err != nil {
			matched = false
			log.Println("error in regular expression of field "+field.Name+" : ", err)
		}
		if !matched {
			setFieldError(field, tag, value, path, fieldname, errors)
			return true
		}
	default:
		log.Println("SanitizeFieldRegexp on type", field.Type.Name(), "not implemented")
		return true
	}
	return false
}

// sanitizeFieldEmail a string
func sanitizeFieldEmail(field reflect.StructField, tag Tag, value reflect.Value, path string, fieldname string, errors *[]FieldError) bool {
	switch field.Type.Kind() {
	case reflect.String:
		matched, err := regexp.MatchString(`^[a-zA-Z0-9.!#$%&'*+/=?^_`+"`"+`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`, value.String())
		if err != nil {
			matched = false
			log.Println("error in regular expression of field "+field.Name+" : ", err)
		}
		if !matched {
			setFieldError(field, tag, value, path, fieldname, errors)
			return true
		}
	default:
		log.Println("SanitizeFieldEmail on type", field.Type.Name(), "not implemented")
		return true
	}
	return false
}

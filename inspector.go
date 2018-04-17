package combi

import (
	"errors"
	"log"
	"reflect"
	"strings"
	// "github.com/davecgh/go-spew/spew"
)

type ObjectData struct {
	Index      int
	FieldInfos []*FieldInfo
}

// FieldInfo to store data about a given struct field
type FieldInfo struct {
	Index      int
	Required   bool
	Zero       bool
	Namespace  string
	Name       string
	Hint       string
	SFlag      string
	LFlag      string
	Type       reflect.Type
	Tags       reflect.StructTag
	Field      reflect.Value
	FieldPtr   interface{}
	Value      interface{}
	FieldInfos []*FieldInfo
}

var (
	ErrStructPtrExpected = errors.New("pointer to struct expected")
	ErrPtrExpected       = errors.New("pointer expected")
)

func InspectStruct(obj interface{}) ([]*FieldInfo, error) {

	// validate is pointer
	ptrVal := reflect.ValueOf(obj)
	if ptrVal.Kind() != reflect.Ptr {
		return nil, ErrStructPtrExpected
	}

	// reflect the underlying value and validate is struct
	val := reflect.Indirect(ptrVal)
	if val.Kind() != reflect.Struct {
		return nil, ErrStructPtrExpected
	}

	// set up index and slice refs
	index := 0
	fieldInfos := []*FieldInfo{}

	// inspect fields (recursive)
	err := inspect(&fieldInfos, &index, "", reflect.StructField{}, ptrVal)
	if err != nil {
		return nil, err
	}

	// fmt.Println("===============================================================")
	// spew.Dump(fieldInfos)
	// fmt.Println("===============================================================")

	return fieldInfos, nil
}

var excludeFieldNames = map[string]struct{}{
	"XMLName": struct{}{},
	"Local":   struct{}{},
	"Space":   struct{}{},
}

/*
inspect uses reflection to pull field data from the underlying value,
intended to analyse command request / response objects
@TODO needs further development - not bulletproof
*/
func inspect(fields *[]*FieldInfo, index *int, ns string, typeField reflect.StructField, ptrVal reflect.Value) error {

	// first check we have a pointer, otherwise we cant dereference & set value
	if ptrVal.Kind() != reflect.Ptr {
		log.Println(ptrVal.Kind())
		return errors.New("reflection to set value on non-pointer")
	}

	val := reflect.Indirect(ptrVal)

	switch val.Kind() {
	case reflect.Int, reflect.String, reflect.Bool:

		// ignore XMLName and other internals
		var ok bool
		if _, ok = excludeFieldNames[typeField.Name]; ok {
			return nil
		}

		// collect field info
		field := &FieldInfo{}
		field.FieldPtr = ptrVal.Interface()
		field.Index = *index
		field.Namespace = ns
		field.Name = typeField.Name
		field.Field = val
		field.Type = val.Type()
		field.Tags = typeField.Tag
		validTag := field.Tags.Get("valid")
		field.SFlag = field.Tags.Get("sFlag")
		field.LFlag = field.Tags.Get("lFlag")
		field.Hint = field.Tags.Get("hint")
		// log.Println(field.Hint)

		// check if field is marked as required
		if validTag != "" {
			if strings.Index(validTag, "required") == 0 {
				field.Required = true
			}
		}

		// check if the value is the zero value for its type
		field.Value = val.Interface()
		if field.Value == reflect.Zero(reflect.TypeOf(field.Value)).Interface() {
			field.Zero = true
		}

		if field.Type == nil {
			log.Fatal("boo")
		}

		// set field then inc index for next
		*fields = append(*fields, field)
		*index++

		return nil

	case reflect.Struct:

		// Iterate over the struct fields and call recursively
		for i := 0; i < val.NumField(); i++ {
			newNS := ""

			// fieldName := val.Type().Field(i).Name
			fieldName := typeField.Name
			if typeField.Name != "" {
				newNS = ns + fieldName + "->"
			}

			err := inspect(fields, index, newNS, val.Type().Field(i), val.Field(i).Addr())
			if err != nil {
				return err
			}
		}

		return nil

	default:
		return errors.New("Unsupported kind: " + val.Kind().String())
	}
}

func splitRequiredFields(fis []*FieldInfo) (required, optional []*FieldInfo) {

	// identify all required and optional values (including nested)
	for _, fi := range fis {
		if len(fi.FieldInfos) > 0 {
			r, o := splitRequiredFields(fi.FieldInfos)
			required = append(required, r...)
			optional = append(optional, o...)
		} else {
			if fi.Required {
				required = append(required, fi)
			} else {
				optional = append(optional, fi)
			}
		}
	}

	return required, optional
}

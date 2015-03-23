package h

import (
	"fmt"
	"html"
	"io"
	"reflect"
	"strconv"
)

type Attributes map[string]interface{}

// Render an attribute value.
func writeValue(w io.Writer, i interface{}) (int, error) {
	var res string
	value := reflect.ValueOf(i)
	switch value.Kind() {
	case reflect.Bool:
		res = strconv.FormatBool(value.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		res = strconv.FormatInt(value.Int(), 10)
	case reflect.Float32, reflect.Float64:
		res = strconv.FormatFloat(value.Float(), 'E', 3, 64)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		res = strconv.FormatUint(value.Uint(), 10)
	case reflect.String:
		res = value.String()
	default:
		return 0, fmt.Errorf(
			`Could not write attribute value "%v" with kind %s`, i, value.Kind())
	}
	return fmt.Fprint(w, html.EscapeString(res))
}

// Check if a value is empty.
func isZero(i interface{}) (bool, error) {
	value := reflect.ValueOf(i)
	switch value.Kind() {
	case reflect.Bool:
		return value.Bool() == false, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0, nil
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint() == 0, nil
	case reflect.String:
		return value.String() == "", nil
	default:
		return false, fmt.Errorf(
			`Could not work with attribute value "%v" with kind %s`, i, value.Kind())
	}
}

// Render a attribute value pair.
func writeKeyValue(w io.Writer, key string, val interface{}) (int, error) {
	var err error
	var skip bool
	skip, err = isZero(val)
	if err != nil {
		return 0, err
	}
	if skip {
		return 0, nil
	}
	var written, i int
	i, err = fmt.Fprintf(w, ` %s`, key)
	written += i
	if err != nil {
		return written, err
	}
	// bool values are not written, only the key is
	if reflect.ValueOf(val).Kind() == reflect.Bool {
		return written, nil
	}
	i, err = fmt.Fprintf(w, `="`)
	written += i
	if err != nil {
		return written, err
	}
	i, err = writeValue(w, val)
	written += i
	if err != nil {
		return written, err
	}
	i, err = fmt.Fprint(w, `"`)
	written += i
	if err != nil {
		return written, err
	}
	return written, nil
}

// Render attributes with using the optional key prefix.
func (attrs Attributes) Write(w io.Writer, prefix string) (int, error) {
	var written, i int
	var err error
	for key, val := range attrs {
		i, err = writeKeyValue(w, prefix+key, val)
		written += i
		if err != nil {
			return written, err
		}
	}
	return written, nil
}
